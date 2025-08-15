use super::PeerIndexTable;
use super::RibEntry;
use anyhow::Result;
use byteorder::{BigEndian, ReadBytesExt};
use chrono::DateTime;
use serde::Serialize;
use serde_with::serde_as;
use std::fmt::{self, Display};
use std::io::Read;
use std::net::Ipv4Addr;

#[serde_as]
#[derive(Debug, Serialize)]
#[allow(dead_code)]
pub struct RibIpV4Unicast {
    time: DateTime<chrono::Utc>,
    sequence_number: u32,       // Sequence number of the RIB entry
    prefix_len: u8,             // Length of the prefix
    prefix: Ipv4Addr,           // network prefix
    entry_count: u16,           // Number of entries in the RIB
    rib_entries: Vec<RibEntry>, // Rib entries
}

impl RibIpV4Unicast {
    pub fn from_reader<R: Read>(
        reader: &mut R,
        peer_index_table: &PeerIndexTable,
        time: DateTime<chrono::Utc>,
    ) -> Result<Self> {
        let sequence_number = reader.read_u32::<BigEndian>()?;
        let prefix_len = reader.read_u8()?;
        // Read the prefix for IPv4 addresses
        // Compute Prefix length in bytes
        let prefix_len_bytes = prefix_len.div_ceil(8);
        let mut prefix_bytes = [0u8; 4];
        reader.read_exact(&mut prefix_bytes[..prefix_len_bytes as usize])?;
        let prefix = Ipv4Addr::from(prefix_bytes);
        // Read the number of entries
        let entry_count = reader.read_u16::<BigEndian>()?;
        println!(
            "sequence_number {}, prefix_len {}, prefix_len_bytes: {}, prefix {}/{}, num_entries {}",
            sequence_number, prefix_len, prefix_len_bytes, prefix, prefix_len, entry_count
        );
        // read the rib entry
        let mut rib_entries: Vec<RibEntry> = Vec::with_capacity(entry_count as usize);
        for _ in 0..entry_count {
            rib_entries.push(RibEntry::from_reader(reader, &peer_index_table)?);
        }
        Ok(RibIpV4Unicast {
            time: time,
            sequence_number,
            prefix_len,
            prefix,
            entry_count,
            rib_entries,
        })
    }
}

impl Display for RibIpV4Unicast {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        for entry in &self.rib_entries {
            writeln!(f, "TIME: {}", self.time.format("%Y-%m-%d %H:%M:%S"))?;
            writeln!(f, "TYPE: TABLE_DUMP_V2/IPV4_UNICAST")?;
            writeln!(f, "PREFIX: {:?}/{:?}", self.prefix, self.prefix_len)?;
            writeln!(f, "SEQUENCE: {}", self.sequence_number)?;
            writeln!(f, "FROM: {:?} AS{:?}", entry.peer_ip, entry.peer_asn)?;
            writeln!(
                f,
                "ORIGINATED: {}",
                entry.originated_time.format("%Y-%m-%d %H:%M:%S")
            )?;
            if let Some(origin) = &entry.bgp_origin {
                writeln!(f, "ORIGIN: {}", origin)?;
            }
            if let Some(as_path) = &entry.bgp_as_path {
                writeln!(
                    f,
                    "ASPATH: {}",
                    as_path
                        .segments
                        .iter()
                        .map(|seg| seg.to_string())
                        .collect::<Vec<_>>()
                        .join(" ")
                )?;
            }
            if let Some(next_hop) = &entry.bgp_next_hop {
                writeln!(f, "NEXT_HOP: {}", next_hop.0)?;
            }
            if let Some(multi_exit_disc) = &entry.bgp_multi_exit_disc {
                writeln!(f, "MULTI_EXIT_DISC: {}", multi_exit_disc.0)?;
            }
            if let Some(communities) = &entry.bgp_community {
                writeln!(
                    f,
                    "COMMUNITIES: {}",
                    communities
                        .0
                        .iter()
                        .map(|(asn, local)| format!("{}:{}", asn, local))
                        .collect::<Vec<_>>()
                        .join(" ")
                )?;
            }
            if let Some(communities) = &entry.bgp_large_community {
                writeln!(
                    f,
                    "LARGE_COMMUNITY: {}",
                    communities
                        .0
                        .iter()
                        .map(|(asn, local, global)| format!("{}:{}:{}", asn, local, global))
                        .collect::<Vec<_>>()
                        .join(" ")
                )?;
            }
            writeln!(f)?;
        }
        Ok(())
    }
}
