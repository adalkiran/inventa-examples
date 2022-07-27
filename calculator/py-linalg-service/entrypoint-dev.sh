echo "Downloading dependent Python modules..."
pip install -r ./requirements.txt

echo "Running into Waiting loop..."
tail -f /dev/null