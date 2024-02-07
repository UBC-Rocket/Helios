import sys
from pathlib import Path

# Stupid hack for protobuf imports to work
sys.path.append(str(Path(Path(__file__).parent, "protocol", "generated")))
