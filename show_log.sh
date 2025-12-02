#!/bin/sh

sqlite3 cronlog-store.db "SELECT * FROM OPRESULTS"
