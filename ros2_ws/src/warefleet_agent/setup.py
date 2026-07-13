from setuptools import find_packages, setup

package_name = 'warefleet_agent'

setup(
    name=package_name,
    version='0.1.0',
    packages=find_packages(exclude=['test']),
    data_files=[
        ('share/ament_index/resource_index/packages',
         ['resource/' + package_name]),
        ('share/' + package_name, ['package.xml']),
    ],
    install_requires=['setuptools'],
    zip_safe=True,
    maintainer='Your Name',
    maintainer_email='you@example.com',
    description='Per-robot task-execution agent and order feeder for WareFleet.',
    license='MIT',
    tests_require=['pytest'],
    entry_points={
        'console_scripts': [
            'agent_node = warefleet_agent.agent_node:main',
            'order_feeder = warefleet_agent.order_feeder:main',
            'mqtt_bridge = warefleet_agent.mqtt_bridge:main',
        ],
    },
)
