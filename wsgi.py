# WSGI configuration for PythonAnywhere
# NOTE: This file is a TEMPLATE - edit the WSGI file directly in PythonAnywhere's Web tab
# Do not put your actual username here in the git repo

import sys
import os

# CHANGE 'YOURUSERNAME' to your PythonAnywhere username
project_home = '/home/YOURUSERNAME/go-calendar'
if project_home not in sys.path:
    sys.path.insert(0, project_home)

os.chdir(project_home)

from app import app as application
