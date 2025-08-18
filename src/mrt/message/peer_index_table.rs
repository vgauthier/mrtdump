use crate::mrt::Error;
use byteorder::{BigEndian, ReadBytesExt};
use std::io::Read;
use std::net::{IpAddr, IpAddr::V4, IpAddr::V6, Ipv4Addr, Ipv6Addr};

#[derive(Debug)]
#[allow(dead_code)]
pub struct PeerEntry {
    pub bgp_id: u32,     // BGP ID of the peer
    pub peer_ip: IpAddr, // IP address of the peer
    pub peer_asn: u32,   // Autonomous System Number of the peer
}

#[derive(Debug)]
#[allow(dead_code)]
pub struct PeerIndexTable {
    pub collector_bgp_id: u32, // BGP ID of the collector
    pub view_name_len: u16,    // Length of the view name
    pub view_name: String,     // Name of the view
    pub nentries: u16,         // Number of entries in the peer index
    pub entries: Vec<PeerEntry>,
}

impl PeerIndexTable {
    pub fn from_reader<R: Read>(reader: &mut R) -> Result<Self, Error> {
        let collector_bgp_id = reader.read_u32::<BigEndian>()?;
        let view_name_len = reader.read_u16::<BigEndian>()?;
        let mut view_name = vec![0u8; view_name_len as usize];
        reader.read_exact(&mut view_name)?;
        let view_name = String::from_utf8(view_name)?;
        let nentries = reader.read_u16::<BigEndian>()?;
        // Read the peer entries
        let mut entries = Vec::with_capacity(nentries.into());
        for _ in 0..nentries {
            let entry = PeerEntry::from_reader(reader)?;
            entries.push(entry);
        }
        Ok(PeerIndexTable {
            collector_bgp_id,
            view_name_len,
            view_name,
            nentries,
            entries,
        })
    }
}

impl PeerEntry {
    pub fn from_reader<R: Read>(reader: &mut R) -> Result<Self, Error> {
        let peer_type = reader.read_u8()?;
        let is_ipv6: u8 = 0x01;
        let is_asn_size_32: u8 = 0x02;
        //bgp_id
        let bgp_id = reader.read_u32::<BigEndian>()?;
        //peer_ip
        let peer_ip: IpAddr = if peer_type & is_ipv6 == is_ipv6 {
            V6(Ipv6Addr::from_bits(reader.read_u128::<BigEndian>()?))
        } else {
            V4(Ipv4Addr::from_bits(reader.read_u32::<BigEndian>()?))
        };
        // peer_asn
        let peer_asn = if peer_type & is_asn_size_32 == is_asn_size_32 {
            // ASN is 32 bits
            reader.read_u32::<BigEndian>()?
        } else {
            // ASN is 16 bits, read as 0 for compatibility
            reader.read_u16::<BigEndian>()? as u32
        };
        Ok(PeerEntry {
            bgp_id,
            peer_ip,
            peer_asn,
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Cursor;

    #[test]
    fn test_peer_index_table_from_reader() {
        let mut cursor = Cursor::new(vec![
            0, 0, 0, 1, // collector_bgp_id
            0, 4, // view_name_len
            b't', b'e', b's', b't', // view_name
            0, 0x01, // nentries
            0x02, // peer_type (IPv4, ASN 16 bits)
            0x00, 0x00, 0x00, 0x02, // bgp_id
            192, 168, 1, 1, // peer_ip (IPv4)
            0x00, 0x00, 0x00, 0x00, // peer_asn (ASN 16 bits)
        ]);
        let table = PeerIndexTable::from_reader(&mut cursor);
        assert!(table.is_ok());
    }
}
