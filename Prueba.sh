#!/bin/bash


Usr=redblu_administrator
Pas=4dm1n1str4t0r
Host=localhost
Port=5672
VHost=vh_business
Cola=qu.cm.events
NumMsgs=800
DirCola=qu.cm.events

Programa="./ExtraeMsgsRabbit"


${Programa} -full -uri="amqp://${Usr}:${Pas}@${Host}:${Port}/${VHost}" -queue="${Cola}" -max-messages=${NumMsgs} -output-dir="${DirCola}"

