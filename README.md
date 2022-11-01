# Websocket
This project is a Websocket broadcaster
---
Following API's are used in the project
1. `/publish`: http POST endpoint: sending messages to this endpoint will broadcast to all websocket clients
```
curl --location --request POST 'http://localhost:8080/publish' \
--header 'Content-Type: application/json' \
--data-raw '{
    "message": "Hey there"
}'
```
---
2. `/ws` : websocket endpoint: Messages incoming from above `publish` api will be broadcasted to every websocket client.
---
