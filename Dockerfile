# Gunakan image Go resmi
FROM golang:1.22-alpine

# Set working directory di dalam container
WORKDIR /app

# Copy file go.mod dan go.sum terlebih dahulu
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy semua file project ke dalam container
COPY . .

# Build aplikasi Go
RUN go build -o main .

# Expose port 8080 untuk Railway
EXPOSE 8080

# Jalankan aplikasi
CMD ["./main"]
