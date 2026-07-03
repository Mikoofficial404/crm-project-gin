#!/bin/bash

COMMAND=$1

case "$COMMAND" in
    crm-start)
        echo "Menjalankan container CRM (Postgres & Redis)..."
        docker-compose up -d
        ;;
    crm-stop)
        echo "Menghentikan container CRM..."
        docker-compose down
        ;;
    crm-restart)
        echo "Merestart container CRM..."
        docker-compose down
        docker-compose up -d
        ;;
    crm-logs)
        echo "Menampilkan logs container CRM..."
        docker-compose logs -f
        ;;
    crm-status)
        echo "Status container CRM:"
        docker-compose ps
        ;;
    *)
        echo "Penggunaan: $0 {crm-start|crm-stop|crm-restart|crm-logs|crm-status}"
        exit 1
esac
