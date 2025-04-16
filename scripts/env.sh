#!/bin/bash

export PGPASSWORD='pass123'
export ROOT_USER_PSQL='postgres'
export DB_HOST=$(hostname -I | awk '{print $1}')
export DB_PORT=5432
export DB_TYPE='postgres'
export DB_SSLMODE=disable

export DB_ORDER_SERVICE='order_service'
export USER_ORDER_SERVICE='order_service'
export PASSWORD_ORDER_SERVICE='aeva0lah0eejaiphaiPhie'
export CGO_ENABLED=1
export JWT_SECRET_KEY='kahxein2Theey2Jae8Doh1'
