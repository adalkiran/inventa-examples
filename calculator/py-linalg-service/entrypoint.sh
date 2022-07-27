echo "Downloading dependent Python modules..."
pip install -r ./requirements.txt

echo "Running application..."
cd src
python app.py
tail -f /dev/null