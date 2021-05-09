FROM python:3.9-buster

RUN apt-get update && apt-get -y install cron

COPY release/linux/amd64/server /app/
COPY *.py /app/
COPY requirements.txt /app
COPY run.sh /app
WORKDIR /app

RUN chmod a+x run.sh
RUN pip3 install -r requirements.txt

COPY cron_job /etc/cron.d/cron_job

RUN chmod 0644 /etc/cron.d/cron_job

RUN crontab /etc/cron.d/cron_job

EXPOSE 80

ENTRYPOINT ["./run.sh"]
