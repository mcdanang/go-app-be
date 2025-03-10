# use official Golang imag
FROM golang:1.16.3-alpine3.13

# set working directory
WORKDIR /app

# copy the source code
COPY . .

# download and install the dependencies
RUN go mod tidy

# build the go app
RUN go build -o api .

# expose the port
EXPOSE 8000

# run the executable
CMD ["./api"] 