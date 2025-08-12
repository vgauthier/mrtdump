# MRTDump

## Introduction 
The Rust implementation of bgpdump exports the binary MRT format to a human-readable format. The MRT format is used to export routing protocol messages, state changes, and contents of the routing information base. It is defined in [RFC6396](https://doi.org/10.17487/rfc6396).

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

## TODOs
* Remove the PeerIndex from the rib entry struct and the String from the ribentry.go
* Implement CVS print option
* Change the function ToJSON print from a struct to map
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

