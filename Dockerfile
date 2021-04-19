FROM golang:1.14-alpine

WORKDIR /go/src/github.com/cronosgmbh/office-checkin-backend
COPY . .

RUN apk update
RUN apk add git
RUN go get -d -v ./...
RUN go install -v ./...

ENV GOOGLE_APPLICATION_CREDENTIALS /go/src/github.com/cronosgmbh/office-checkin-backend/firebase.json
ENV GIN_MODE release
ENV CRONOS_MONGO_HOST localhost
ENV CRONOS_MONGO_USERNAME root
ENV CRONOS_MONGO_PASSWORD example
ENV CRONOS_MONGO_DB office-checkin

EXPOSE 3000

CMD ["office-checkin-backend"]
