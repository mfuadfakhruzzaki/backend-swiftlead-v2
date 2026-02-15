import argparse
import json
import time
import os
import random
import uuid
import sys

# Try to import paho.mqtt.client
try:
    import paho.mqtt.client as mqtt
except ImportError:
    print("Error: 'paho-mqtt' library is required. Install it using: pip install paho-mqtt")
    sys.exit(1)

# Default configuration (can be overridden by args or env)
DEFAULT_BROKER = os.getenv("MQTT_BROKER", "localhost")
DEFAULT_PORT = int(os.getenv("MQTT_PORT", "1883"))
DEFAULT_TOPIC_PREFIX = os.getenv("MQTT_TOPIC_PREFIX", "swiftlead/tel")
DEFAULT_CLIENT_ID = f"simulator-{uuid.uuid4().hex[:8]}"

def main():
    parser = argparse.ArgumentParser(description="Simulate a SwiftLead Node sending telemetry to update status to Online")
    parser.add_argument("--broker", default=DEFAULT_BROKER, help="MQTT Broker address")
    parser.add_argument("--port", type=int, default=DEFAULT_PORT, help="MQTT Broker port")
    parser.add_argument("--topic-prefix", default=DEFAULT_TOPIC_PREFIX, help="MQTT Topic prefix")
    parser.add_argument("--esp32-uid", required=True, help="ESP32 UID (Mac Address or Unique ID) of the node")
    parser.add_argument("--temp", type=float, default=28.0, help="Temperature value")
    parser.add_argument("--rh", type=float, default=70.0, help="Humidity value")
    parser.add_argument("--nh3", type=float, default=10.0, help="Ammonia value")
    parser.add_argument("--username", help="MQTT Username")
    parser.add_argument("--password", help="MQTT Password")
    parser.add_argument("--loop", action="store_true", help="Send data continuously every 5 seconds")
    
    args = parser.parse_args()

    client = mqtt.Client(client_id=DEFAULT_CLIENT_ID)
    
    if args.username and args.password:
        client.username_pw_set(args.username, args.password)

    try:
        print(f"Connecting to MQTT Broker at {args.broker}:{args.port}...")
        client.connect(args.broker, args.port, 60)
        client.loop_start()
    except Exception as e:
        print(f"Failed to connect: {e}")
        sys.exit(1)

    topic = f"{args.topic_prefix}/{args.esp32_uid}"
    
    # Construct payload matching models.SensorPayload
    payload = {
        "esp32_uid": args.esp32_uid,
        "temp": args.temp,
        "rh": args.rh,
        "nh3": args.nh3,
        "rssi": -random.randint(40, 90),
        "timestamp": int(time.time()),
        "seq": 1
    }

    def send_telemetry():
        payload["timestamp"] = int(time.time())
        payload["temp"] = args.temp + random.uniform(-0.5, 0.5)
        payload["rh"] = args.rh + random.uniform(-1.0, 1.0)
        
        json_payload = json.dumps(payload)
        info = client.publish(topic, json_payload)
        info.wait_for_publish()
        print(f"Published to {topic}: {json_payload}")
        print("Node status should now be ONLINE.")

    send_telemetry()

    if args.loop:
        try:
            while True:
                time.sleep(5)
                payload["seq"] += 1
                send_telemetry()
        except KeyboardInterrupt:
            print("\nStopping...")
    
    client.loop_stop()
    client.disconnect()

if __name__ == "__main__":
    main()
