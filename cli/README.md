try gobgp with CLI first

```
go run main.go

gobgp global
AS:        0
Router-ID:
------
gobgp global as 64512 router-id 10.0.0.1 listen-port 179

gobgp global

AS:        64512
Router-ID: 10.0.0.1
Listening Port: 179, Addresses: 0.0.0.0, ::
-------
gobgp neighbor add 192.168.2.1 as 64513
(⎈|N/A:N/A)~/repo/liyi/gobgp gobgp neighbor
Peer           AS  Up/Down State       |#Received  Accepted
192.168.2.1 64513 00:00:24 Establ      |        2         2
(⎈|N/A:N/A)~/repo/liyi/gobgp gobgp global rib
   Network              Next Hop             AS_PATH              Age        Attrs
*> 24.57.254.0/24       192.168.2.1          64513                00:00:03   [{Origin: ?}]
*> 192.168.2.0/24       192.168.2.1          64513                00:00:03   [{Origin: ?}]
-------
gobgp global rib add 192.168.1.0/24 -a ipv4 
gobgp global rib

   Network              Next Hop             AS_PATH              Age        Attrs
*> 24.57.254.0/24       192.168.2.1          64513                00:02:35   [{Origin: ?}]
*> 192.168.1.0/24       0.0.0.0                                   00:00:12   [{Origin: ?}]
*> 192.168.2.0/24       192.168.2.1          64513                00:02:35   [{Origin: ?}]

```