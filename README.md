1. mysql
   docker run -itd -p 3306:3306 --net mynet --ip 172.18.0.10 -e MYSQL_ROOT_PASSWORD=root -v C:\develop\docker\mysql:/var/lib/mysql --name mysql mysql:5.7


2. influxdb
   docker run -itd --name=influxdb --net mynet --ip 172.18.0.11 -p 8086:8086 -p 8083:8083 -v C:\develop\docker\influxdb:/var/lib/influxdb influxdb:1.8.9


show retention policies
create database monitor
create retention policy "expire_half_year" on "monitor" duration 25w replication 1 default
CREATE RETENTION POLICY "expire_30d" ON "monitor" DURATION 30d REPLICATION 1


3. grafana
   docker run -itd --name=grafana --net mynet --ip 172.18.0.12 -p 3000:3000  grafana/grafana:8.1.3


4. xxl
   docker run -it --net mynet --ip 172.18.0.13 -e PARAMS="--spring.datasource.url=jdbc:mysql://mysql:3306/xxl_job?Unicode=true&characterEncoding=UTF-8 --spring.datasource.username=root --spring.datasource.password=root" -p 8080:8080 --name xxl --restart=always  -d xuxueli/xxl-job-admin:2.3.0


5. spider
   docker build -t spider .
   docker run -itd -p 8081:8081 --net mynet --ip 172.18.0.14 --name spider spider


6. spider_admin
   docker build -t spider_admin .
   docker run -itd --net mynet --ip 172.18.0.15 -p 8082:80 -e NGINX_UPSTREAM="server 172.18.0.14:8081;" --name spider_admin spider_admin