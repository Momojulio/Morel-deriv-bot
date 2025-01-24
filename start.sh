docker build -t deriv-bot . --no-cache
docker run --publish 80:80 --restart always -d deriv-bot