# MRTDump

## Introduction 
The Rust implementation of bgpdump to exports the binary MRT format to a human-readable format. The MRT format is used to export routing protocol messages, state changes, and contents of the routing information base. It is defined in [RFC6396](https://doi.org/10.17487/rfc6396).

### Currently supported MRT types

| Name            | Value | Is Implemented  |
|-----------------|-------|-----------------|
| TABLE\_DUMP     | 12    | No              |
| TABLE\_DUMP\_V2 | 13    | Yes             |
| BGP4MP          | 16    | No              |
| BGP4MP\_ET      | 17    | No              |
| ISIS            | 32    | No              |
| ISIS_ET         | 33    | No              |
| OSPFv3          | 48    | No              |  
| OSPFv3_ET       | 49    | No              |

# Usage

```bash 
> mrtdump --help
Read MRT binary files and format and print it in a human-readable format JSON/CSV/PRINT

Usage: mrtdump [OPTIONS] <INPUT_FILE>

Arguments:
  <INPUT_FILE>  Input file path MRT format

Options:
  -p, --print    Multi-line, human-readable (the default)
  -j, --json     Output in JSON format
  -h, --help     Print help
  -V, --version  Print version
> mrtdump rib.20250701.0000
TIME: 2025-07-01 00:00:00
TYPE: TABLE_DUMP_V2/IPV4_UNICAST
PREFIX: 0.0.0.0/0
SEQUENCE: 0
FROM: 87.121.64.4 AS57463
ORIGINATED: 2025-06-26 21:10:33
ORIGIN: IGP
ASPATH: 57463 3356
NEXT_HOP: 87.121.64.4
COMMUNITIES: 1:1085 64700:3356 65400:1 65400:65500
LARGE_COMMUNITY: 57463:64700:3356

TIME: 2025-07-01 00:00:00
TYPE: TABLE_DUMP_V2/IPV4_UNICAST
PREFIX: 0.0.0.0/0
SEQUENCE: 0
FROM: 94.156.252.18 AS34224
ORIGINATED: 2025-06-30 17:21:08
ORIGIN: IGP
ASPATH: 34224 3356
NEXT_HOP: 94.156.252.18
MULTI_EXIT_DISC: 0
COMMUNITIES: 34224:333
...
```

## TODOs
* Better error handling
* Implement CVS print option
* Implement a read from bzip, zip and gz file
* Add option to output to a file



## Licence
Licensed under the Apache License, Version 2.0
Copyright (C) 2025 Vincent Gauthier

## References 

1. [Github ipgiri: An implementation of IPv4 Routing Lookup Table and MRT files parsing in python.](https://github.com/gabhijit/ipgiri/)
2. [Github MRTParse](https://github.com/t2mune/mrtparse/)
3. Rekhter, Y., Li, T., & Hares, S. (Eds.). (2006). A Border Gateway Protocol 4 (BGP-4). RFC Editor. [https://doi.org/10.17487/rfc4271](https://doi.org/10.17487/rfc4271)
4. Blunk, L., Karir, M., & Labovitz, C. (2011). Multi-Threaded Routing Toolkit (MRT) Routing Information Export Format. RFC Editor. [https://doi.org/10.17487/rfc6396](https://doi.org/10.17487/rfc6396)

