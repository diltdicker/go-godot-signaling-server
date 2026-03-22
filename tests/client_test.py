import asyncio
import websockets

async def connect_to_go():
    # Note the protocol prefix 'ws://'
    uri = "ws://localhost:8080/ws"
    
    try:
        # 'async with' ensures the connection is closed automatically
        async with websockets.connect(uri) as websocket:
            print(f"Connected to {uri}")

            # Send a simple text message
            message = "Hello from Python!"
            await websocket.send(message)
            print(f"> Sent: {message}")

            # Wait for the server to echo it back
            response = await websocket.recv()
            print(f"< Received: {response}")

    except ConnectionRefusedError:
        print("Error: Go server is not running on localhost:8080")
    except Exception as e:
        print(f"An unexpected error occurred: {e}")

if __name__ == "__main__":
    asyncio.run(connect_to_go())