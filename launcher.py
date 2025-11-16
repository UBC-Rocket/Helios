import docker
import os
import sys
import time
import threading

from google.protobuf import json_format
import generated.python.config.component_pb2 as component

# Prune leftover stopped containers from build step
os.environ["DOCKER_BUILDKIT"] = "1"

build_threads = []

LOCAL = os.path.dirname(sys.executable)

client = docker.from_env()
images = {
  "Helios": ('./', 'helios:latest'),
  # "RocketDecoder": ('../RocketDecoder', 'rocketdecoder:latest'),
  # #"UI": ('../UI', 'ui:latest'),
  # "Livestream": ('../Livestream', 'livestream:latest')
}

volumes_config = {
  '/var/run/docker.sock': {
    'bind': '/var/run/docker.sock',
    'mode': 'rw' # Read-write access
  }
}

with open('./runtime_hash.txt', 'r') as runtime_hash_file:
  RUNTIME_HASH = f"{runtime_hash_file.readline().strip()}-{time.time()}"

# Have logic for changing or using current runtime hash here

def main():
  print('Starting launcher with hash:', RUNTIME_HASH)
  print('Building Docker images...')  
  build_images(client, images)

  tree, path = generateComponentTree()
  print('Generated component tree:', json_format.MessageToJson(tree))
  
  print('Starting Helios container...')
  StartHelios(path)


def StartHelios(tree_path = None):
  Helios = images["Helios"]
  HeliosContainer = None

  # Check if there is an existing Helios container
  # Either remove or restart it based on hash comparison
  existing_containers = client.containers.list(all=True, filters={"name": "Helios"})
  if existing_containers:
    existing_container = existing_containers[0]
    existing_hash = existing_container.labels.get('runtime_hash', None)
    
    if existing_hash == RUNTIME_HASH:
      if existing_container.status == "running":
        print("Helios container is already running. No additional builds will be run.")
        return
      elif existing_container.status == "exited":
        print("Helios container is up-to-date. Starting it back up...")

        existing_container.restart()
        HeliosContainer = existing_container
        return
    else:
      print("Helios container is outdated. Removing it...")
      existing_container.remove(force=True)

  # Create the Helios container if not found or removed
  if HeliosContainer is None:
    print("Starting a new Helios container...")
    HeliosContainer = client.containers.run(
      Helios[1], 
      name='Helios', 
      detach=True,
      volumes=volumes_config, # Give the container access to docker.sock
      labels={
        'runtime_hash': RUNTIME_HASH
      },
      environment={
        'RUNTIME_HASH': RUNTIME_HASH,
        'COMPONENT_TREE_PATH': tree_path if tree_path else ''
      }
    )

  #TODO: Send the component tree and images over to Helios


def build_images(client, images):
  try:
    for image in images:
      (build_context_path, image_tag) = images[image]

      thread = threading.Thread(target=_build_image, args=(client, build_context_path, image_tag))
      thread.start()

      build_threads.append(thread)

  except docker.errors.BuildError as e:
    print(f"Error building image: {e}")
    for line in e.build_log:
      if 'stream' in line:
          print(line['stream'].strip())
  except Exception as e:
    print(f"An unexpected error occurred: {e}")

  # Wait for all threads to finish
  for i in range(len(build_threads)):
    build_threads[i].join()

def _build_image(client, path, tag):
  start = time.time()
  image, build_logs = client.images.build(path=path, tag=tag, rm=True)

  print(f"Image '{image.tags[0]}' built successfully in {round(time.time() - start, 2)}s.")


def generateComponentTree():
  # Example of generating a component tree using protobuf
  tree_location = "./component_tree.json"

  leaf_component = component.BaseComponent()
  leaf_component.name = "LeafComponent"
  leaf = component.Component()
  leaf.path = "/path/to/leaf"
  leaf.tag = "leaf_tag"
  leaf.id = "leaf_1"
  leaf_component.leaf.CopyFrom(leaf)

  branch_component = component.ComponentGroup()
  branch_component.children.extend([leaf_component])

  root = component.BaseComponent()
  root.name = "RootComponent"
  root.branch.CopyFrom(branch_component)

  component_tree = component.ComponentTree()
  component_tree.root.CopyFrom(root)
  component_tree.version = "1.0.0"

  with open(tree_location, "w") as f:
    f.write(json_format.MessageToJson(component_tree))

  return component_tree, tree_location


if __name__ == "__main__":
  main()