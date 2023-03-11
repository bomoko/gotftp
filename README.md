# goTftp

This is purely an exercise in go development for me. 
I've been wanting to implement TFTP for years - I've had a copy of the RFC printed out and in my stack of papers for ages.
So, now, I'm going to implement it in golang, and maybe try do something interesting with the backend. Perhaps minio.

## Notes while devving

* I've got a `./files` directory here to be used as the source and target for TFTP while devving. Will change that up to take in a directory in the future
* I'd really love to be able to capture packets between actual implementations of TFTP
* Reading the source of https://github.com/vcabbage/go-tftp is useful - I'll probably steal some of the ways of working with the datagrams from there.



## Some notes about tftp and how it works

### Encoding

I set up a local tftpd and then captured some packets

```
07:06:58.283228 IP localhost.33920 > localhost.tftp: TFTP, length 23, RRQ "helloworld.txt" octet
	0x0000:  4500 0033 40b7 4000 4011 fc00 7f00 0001  E..3@.@.@.......
	0x0010:  7f00 0001 8480 0045 001f fe32 0001 6865  .......E...2..he
	0x0020:  6c6c 6f77 6f72 6c64 2e74 7874 006f 6374  lloworld.txt.oct
	0x0030:  6574 00                                  et.
07:06:58.285984 IP localhost.60107 > localhost.33920: TFTP, length 7, DATA block 1
	0x0000:  4500 0023 2d40 0000 4011 4f88 7f00 0001  E..#-@..@.O.....
	0x0010:  7f00 0001 eacb 8480 000f fe22 0003 0001  ..........."....
	0x0020:  6869 0a                                  hi.
07:06:58.286104 IP localhost.33920 > localhost.60107: TFTP, length 4, ACK block 1
	0x0000:  4500 0020 40b8 4000 4011 fc12 7f00 0001  E...@.@.@.......
	0x0010:  7f00 0001 8480 eacb 000c fe1f 0004 0001  ................

```

I was interested in how the details were encoded. 
* In particular, I was worried about how tftp might deal with spaces in filenames - although I think I should just ignore this case, because it's stupid.
* If you examine the captured packets above, you'll see that the ascii output for `.` and the `null` character are the same (`.` in both cases) - but the hex representation is `2e` and `00`