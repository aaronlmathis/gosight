#!/bin/bash

# Build Tailwind CSS
npx @tailwindcss/cli -c tailwind.config.js -i ./css/input.css -o ./css/output.css --watch