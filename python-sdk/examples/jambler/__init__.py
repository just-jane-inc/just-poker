import logging
import sys

logging.basicConfig(
    handlers=[
        logging.FileHandler("jambler.log", mode="a"),
        logging.StreamHandler(sys.stdout),
    ],
    format="%(asctime)s - %(levelname)s - [%(name)s] %(message)s",
    level=logging.INFO,  # Capture INFO, WARNING, ERROR, and CRITICAL
)
