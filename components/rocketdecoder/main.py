import socket
import time

HOST = 'Helios' #TODO: Get this host name and port from initial env variables
PORT = 5000

connected = False
stream = None

print("Hello from rocketdecoder container!")

with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
  while not connected:
    try:
      s.connect((HOST, PORT))
      stream = s
      connected = True
      #s.sendall(b"Hello from client")
      data = s.recv(1024)

      print(f"rocketdecoder - Received from server: {data.decode()}")
    except ConnectionRefusedError:
        print(f"Connection refused. Retrying in 5 seconds...")
        time.sleep(5)
    except socket.gaierror:
        print(f"Hostname resolution failed. Retrying in 5 seconds...")
        time.sleep(5)
    except Exception as e:
        print(f"An unexpected error occurred: {e}. Retrying in 5 seconds...")
        time.sleep(5)

  if connected:
    s.sendall(b"Hello from rocketdecoder through port!~")

    while True: 
      data = s.recv(1024)
      if data:
        print(f"rocketdecoder - Received: {data.decode('utf-8')}")
      else:
        print("Connection closed by peer.")

      time.sleep(5)