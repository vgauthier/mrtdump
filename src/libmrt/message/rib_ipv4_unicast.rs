use super::PeerIndexTable;
use super::RibEntry;
use anyhow::Result;
use byteorder::{BigEndian, ReadBytesExt};
use chrono::DateTime;
use std::fmt::{self, Display};
use std::io::Read;
use std::net::Ipv4Addr;
use std::rc::Rc;

#[derive(Debug)]
#[allow(dead_code)]
pub struct RibIpV4Unicast {
    sequence_number: u32,                 // Sequence number of the RIB entry
    prefix_len: u8,                       // Length of the prefix
    prefix: Ipv4Addr,                     // network prefix
    entry_count: u16,                     // Number of entries in the RIB
    rib_entries: Vec<RibEntry>,           // Rib entries
    peer_index_table: Rc<PeerIndexTable>, // Reference to the PeerIndexTable
}

impl RibIpV4Unicast {
    pub fn from_reader<R: Read>(
        reader: &mut R,
        peer_index_table: Rc<PeerIndexTable>,
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
            rib_entries.push(RibEntry::from_reader(reader)?);
        }
        Ok(RibIpV4Unicast {
            sequence_number,
            prefix_len,
            prefix,
            entry_count,
            rib_entries,
            peer_index_table,
        })
    }
}

impl Display for RibIpV4Unicast {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        for entry in &self.rib_entries {
            writeln!(
                f,
                "TIME: {:?}",
                DateTime::from_timestamp(entry.originated_time as i64, 0)
            )?;
            writeln!(f, "TYPE: TABLE_DUMP_V2/IPV4_UNICAST")?;
            writeln!(f, "PREFIX: {:?}/{:?}", self.prefix, self.prefix_len)?;
            writeln!(f, "SEQUENCE: {}", self.sequence_number)?;
            writeln!(
                f,
                "FROM: {:?} AS{:?}",
                self.peer_index_table.entries[entry.peer_index as usize].peer_ip,
                self.peer_index_table.entries[entry.peer_index as usize].peer_asn
            )?;
            writeln!(
                f,
                "ORIGINATED: {}",
                DateTime::from_timestamp(entry.originated_time as i64, 0)
                    .ok_or(std::fmt::Error)?
                    .format("%Y-%m-%d %H:%M:%S")
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
                writeln!(f, "NEXT_HOP: {}", next_hop.ip)?;
            }
            if let Some(multi_exit_disc) = &entry.bgp_multi_exit_disc {
                writeln!(f, "MULTI_EXIT_DISC: {}", multi_exit_disc.metric)?;
            }
            if let Some(communities) = &entry.bgp_community {
                writeln!(
                    f,
                    "COMMUNITIES: {}",
                    communities
                        .community
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
                        .community
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
