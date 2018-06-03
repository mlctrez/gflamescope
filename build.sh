#!/usr/bin/env bash

mkdir -p temp_build

curl -o temp_build/flamescope.zip https://codeload.github.com/Netflix/flamescope/zip/master

cd temp_build

rm -rf flamescope-master

unzip flamescope.zip

cd flamescope-master/app/public

cp index.html _index.html

zip ../../../flamescope-ui.zip -q -r *

cd ../../../

rm -rf gflamescope

vgo build -o gflamescope ../main.go

cat flamescope-ui.zip >> gflamescope

zip -q -A gflamescope

chmod +x ./gflamescope

cp ./gflamescope ../

cd ../

rm -rf temp_build

#./gflamescope



