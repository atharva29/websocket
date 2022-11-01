docker build -t atharva29/wsserver:latest .
docker push atharva29/wsserver:latest
docker run -p 8080:8080 atharva29/wsserver:latest