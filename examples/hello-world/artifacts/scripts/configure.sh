#!/bin/sh
set -e

NODE_TEMPLATE=$1

if ! grep -qF TOSCA /usr/src/app/views/home.handlebars; then
	echo "<div style=\"font-size:32pt\">Part of TOSCA node: \"$NODE_TEMPLATE\"</div>" >> /usr/src/app/views/home.handlebars
fi
