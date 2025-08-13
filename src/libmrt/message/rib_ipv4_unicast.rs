use super::RibEntry;
use anyhow::Result;
use byteorder::{BigEndian, ReadBytesExt};
use std::io::Read;
use std::net::Ipv4Addr;

#[derive(Debug)]
#[allow(dead_code)]
pub struct RibIpV4Unicast {
    sequence_number: u32, // Sequence number of the RIB entry
    prefix_len: u8,       // Length of the prefix
    prefix: Ipv4Addr,     // network prefix
    entry_count: u16,     // Number of entries in the RIB
    rib_entries: Vec<RibEntry>,
}

impl RibIpV4Unicast {
    pub fn from_reader<R: Read>(reader: &mut R) -> Result<Self> {
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
        })
    }

    pub fn to_string(&self) -> String {
        let mut result = String::new();
        for entry in &self.rib_entries {
            result.push_str(&format!("TIME: {}\n", entry.originated_time));
            result.push_str("TYPE: TABLE_DUMP_V2/IPV4_UNICAST\n");
        }
        result
    }
}
