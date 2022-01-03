# E:D Event Hub

A small demo how to use the watched package. It is a program that watches
ED events and sends them via a lightweight protocol to the standard input
of other processes that are started as so called “plugins”.

It also can send events via TCP/IP steams to listeners implemented
using the `edehnet` sub-package. E.g. see the `edehdump` example.
