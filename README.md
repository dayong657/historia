historia
========

A three phase commit protocol implementation in `go` (`golang`).

For a description of the platform, see the writeup in `docs/`.

Building
--------

The easiest way to build is to run with `make` or just download the full 
repository which contains executables for Windows, Mac and Linux in the `build/`
directory.

To speed up compile times you can select an architecture by running 
`make windows_dist`, `make darwin_dist`, or `make linux_dist`.

Testing
-------

There are two ways of testing the software, one is running an interactive server
the other is the demonstration program.

You can validate the code itself by running `go test ./checkup/ ./threephase/` 
in the project root folder.

To test the actual program, you can run the server generated in the build directory.
The following setup would run three servers on :8000, :8001, and :8002

	./server 1 localhost:8000 localhost:8001 localhost:8002
	./server 2 localhost:8000 localhost:8001 localhost:8002
	./server 3 localhost:8000 localhost:8001 localhost:8002

You can access the servers at `http://localhost:800X/`. The root page will give
information about the items on the server and the status of its peers. If you 
navigate to `http://localhost:800X/log/MYSTRING` it will replicate `MYSTRING` 
across the nodes.

If you want to try slamming the server with requests, you can use the `hammer`
executable:

	./hammer --seconds 1 --threads 5 localhost:8000

This will "hammer" `localhost:8000` with requests for 1 second from 5 threads.
