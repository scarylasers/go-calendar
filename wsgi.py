# WSGI configuration for PythonAnywhere
# Update the path below to match your PythonAnywhere username

import sys
import os

# Add your project directory to the sys.path
# Change 'yourusername' to your actual PythonAnywhere username
project_home = '/home/yourusername/go-calendar'
if project_home not in sys.path:
    sys.path.insert(0, project_home)

# Set the working directory
os.chdir(project_home)

# Import the Flask app
from app import app as application
