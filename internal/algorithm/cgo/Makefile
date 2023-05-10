#Makefile
CC=gcc
CFLAGS=-Wall -O2 -shared
LIB=hello.dll
OBJS=hello.o

all: $(LIB)

$(LIB): $(OBJS)
	$(CC) $(CFLAGS) -o $@ $<

%.o: %.c
	$(CC) $(CFLAGS) -c -o $@ $<

clean:
	rm -f $(OBJS) $(LIB)
