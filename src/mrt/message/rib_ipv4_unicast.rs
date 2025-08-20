use super::PeerIndexTable;
use super::RibEntry;
use crate::mrt::Error;
use byteorder::{BigEndian, ReadBytesExt};
use chrono::DateTime;
use serde::Serialize;
use serde_with::serde_as;
use std::fmt::{self, Display};
use std::io::Read;
use std::net::{IpAddr, Ipv4Addr};

#[serde_as]
#[derive(Debug, Serialize)]
#[allow(dead_code)]
struct CsvRecord {
    record_type: String,
    datetime: DateTime<chrono::Utc>,
    entry_type: String,
    peer_ip: IpAddr,
    peer_asn: u32,
    prefix_with_len: String,
    as_path: String,
    bgp_origin: String,
}

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
    ) -> Result<Self, Error> {
        let sequence_number = reader.read_u32::<BigEndian>()?;
        let prefix_len = reader.read_u8()?;
        // Read the prefix for IPv4 addresses :
        // Compute Prefix length in bytes, the Prefix field contains address
        // prefixes followed by enough trailing bits to make the end of the
        // field fall on an octet boundary
        let prefix_len_bytes = prefix_len.div_ceil(8);
        let mut prefix_bytes = [0u8; 4];
        reader.read_exact(&mut prefix_bytes[..prefix_len_bytes as usize])?;
        let prefix = Ipv4Addr::from(prefix_bytes);
        // Read the number of entries
        let entry_count = reader.read_u16::<BigEndian>()?;
        // read the rib entry
        let mut rib_entries: Vec<RibEntry> = Vec::with_capacity(entry_count.into());
        for _ in 0..entry_count {
            let entry = RibEntry::from_reader(reader, peer_index_table)?;
            rib_entries.push(entry);
        }
        Ok(RibIpV4Unicast {
            time,
            sequence_number,
            prefix_len,
            prefix,
            entry_count,
            rib_entries,
        })
    }

    pub fn write_csv_records<W: std::io::Write>(&self, writer: &mut W) -> Result<(), Error> {
        let mut csv_writer = csv::WriterBuilder::new()
            .delimiter(b'|')
            .has_headers(false)
            .from_writer(writer);
        for entry in &self.rib_entries {
            csv_writer.serialize(CsvRecord {
                record_type: "TABLE_DUMP2".to_string(),
                datetime: self.time,
                entry_type: "B".to_string(),
                peer_ip: entry.peer_ip,
                peer_asn: entry.peer_asn,
                prefix_with_len: format!("{}/{}", self.prefix, self.prefix_len),
                as_path: entry
                    .bgp_as_path
                    .as_ref()
                    .unwrap()
                    .segments
                    .iter()
                    .map(|seg| seg.to_string())
                    .collect::<Vec<_>>()
                    .join(" "),
                bgp_origin: entry.bgp_origin.as_ref().unwrap().to_string(),
            })?;
        }
        csv_writer.flush()?;
        Ok(())
    }

    pub fn write_json_records<W: std::io::Write>(&self, writer: &mut W) -> Result<(), Error> {
        let json = serde_json::to_string_pretty(self)?;
        writeln!(writer, "{}", json)?;
        Ok(())
    }
}

impl Display for RibIpV4Unicast {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        let mut entries = String::new();
        for entry in &self.rib_entries {
            let mut entry_string = format!(
                "TIME: {}\nTYPE: TABLE_DUMP_V2/IPV4_UNICAST\nPREFIX: {}/{}\nSEQUENCE: {}\nFROM: {} AS {}\nORIGINATED: {}\n",
                self.time.format("%Y-%m-%d %H:%M:%S"),
                self.prefix,
                self.prefix_len,
                self.sequence_number,
                entry.peer_ip,
                entry.peer_asn,
                entry.originated_time.format("%Y-%m-%d %H:%M:%S")
            );
            if let Some(origin) = &entry.bgp_origin {
                entry_string.push_str(&format!("ORIGIN: {}\n", origin));
            }
            if let Some(as_path) = &entry.bgp_as_path {
                entry_string.push_str(&format!(
                    "ASPATH: {}\n",
                    as_path
                        .segments
                        .iter()
                        .map(|seg| seg.to_string())
                        .collect::<Vec<_>>()
                        .join(" ")
                ));
            }
            if let Some(next_hop) = &entry.bgp_next_hop {
                entry_string.push_str(&format!("NEXT_HOP: {}\n", next_hop.0));
            }
            if let Some(multi_exit_disc) = &entry.bgp_multi_exit_disc {
                entry_string.push_str(&format!("MULTI_EXIT_DISC: {}\n", multi_exit_disc.0));
            }
            if let Some(communities) = &entry.bgp_community {
                entry_string.push_str(&format!(
                    "COMMUNITIES: {}\n",
                    communities
                        .0
                        .iter()
                        .map(|(asn, local)| format!("{}:{}", asn, local))
                        .collect::<Vec<_>>()
                        .join(" ")
                ));
            }
            if let Some(communities) = &entry.bgp_large_community {
                entry_string.push_str(&format!(
                    "LARGE_COMMUNITY: {}\n",
                    communities
                        .0
                        .iter()
                        .map(|(asn, local, global)| format!("{}:{}:{}", asn, local, global))
                        .collect::<Vec<_>>()
                        .join(" ")
                ));
            }
            if let Some(aggregator) = &entry.bgp_aggregator {
                entry_string.push_str(&format!(
                    "AGGREGATOR: {} {}\n",
                    aggregator.asn, aggregator.ip
                ));
            }
            entries.push_str(&format!("{}\n", entry_string));
        }
        writeln!(f, "{}", entries)?;
        Ok(())
    }
}
