// Command fleet-manager is the WareFleet coordination service.
//
// It wires together: MQTT (robot I/O) -> Coordinator (allocation + MAPF +
// dependability) -> Prometheus metrics. Strategy and mode are configured via
// environment variables so the benchmark can sweep them without recompiling.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourname/warefleet/fleet_manager/internal/allocation"
	"github.com/yourname/warefleet/fleet_manager/internal/coordination"
	"github.com/yourname/warefleet/fleet_manager/internal/metrics"
	"github.com/yourname/warefleet/fleet_manager/internal/mqtt"
)

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	var (
		broker      = env("WAREFLEET_MQTT_BROKER", "tcp://localhost:1883")
		allocName   = env("WAREFLEET_ALLOCATION", "greedy")
		planName    = env("WAREFLEET_COORDINATION", "prioritized")
		mode        = coordination.Mode(env("WAREFLEET_MODE", "centralized"))
		metricsAddr = env("WAREFLEET_METRICS_ADDR", ":2112")
	)

	alloc, err := allocation.New(allocName)
	if err != nil {
		log.Fatalf("allocation: %v", err)
	}
	planner, err := coordination.NewPlanner(planName)
	if err != nil {
		log.Fatalf("planner: %v", err)
	}

	mqttClient := mqtt.New(broker)
	coord := coordination.New(alloc, planner, mqttClient, mode)

	// Serve Prometheus metrics.
	go func() {
		log.Printf("metrics on %s/metrics", metricsAddr)
		if err := metrics.Serve(metricsAddr); err != nil {
			log.Printf("metrics server: %v", err)
		}
	}()

	// Wire MQTT inbound to the coordinator.
	if err := mqttClient.Connect(mqtt.Handlers{
		OnOrder:      coord.AddTask,
		OnRobotState: coord.UpdateRobot,
	}); err != nil {
		log.Fatalf("mqtt connect: %v", err)
	}

	log.Printf("WareFleet fleet-manager up | allocation=%s coordination=%s mode=%s",
		allocName, planName, mode)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	coord.Run(ctx, 500*time.Millisecond)
	log.Println("shutting down")
}
