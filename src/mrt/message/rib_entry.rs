use super::{
    BgpAsPath, BgpAttributeHeader, BgpAttributeType, BgpCommunity, BgpLargeCommunity,
    BgpMultiExitDisc, BgpNextHop, BgpOrigin, PeerIndexTable,
};
use crate::mrt;
use anyhow::Result;
use byteorder::{BigEndian, ReadBytesExt};
use chrono::DateTime;
use core::net;
use serde::Serialize;
use serde_with::skip_serializing_none;
use std::io::Read;

#[skip_serializing_none]
#[derive(Debug, Serialize)]
#[allow(dead_code)]
pub struct RibEntry {
    pub peer_index: u16,
    pub peer_asn: u32,
    pub peer_ip: net::IpAddr,
    pub originated_time: DateTime<chrono::Utc>,
    pub attribute_length: u16,
    pub bgp_origin: Option<BgpOrigin>,
    pub bgp_as_path: Option<BgpAsPath>,
    pub bgp_next_hop: Option<BgpNextHop>,
    pub bgp_community: Option<BgpCommunity>,
    pub bgp_large_community: Option<BgpLargeCommunity>,
    pub bgp_multi_exit_disc: Option<BgpMultiExitDisc>,
}

impl RibEntry {
    pub fn from_reader<R: Read>(reader: &mut R, peer_index_table: &PeerIndexTable) -> Result<Self> {
        let peer_index = reader.read_u16::<BigEndian>()?;
        let originated_time = reader.read_u32::<BigEndian>()?;
        let originated_time =
            DateTime::from_timestamp(originated_time as i64, 0).ok_or(std::fmt::Error)?;
        let attribute_length = reader.read_u16::<BigEndian>()?;
        println!(
            "peer_index {}, originated_time {}, attribute_length {}",
            peer_index, originated_time, attribute_length
        );
        // Here you would typically read the attributes based on the attribute_length
        let mut rib_entry = RibEntry {
            peer_index,
            originated_time,
            attribute_length,
            peer_asn: peer_index_table.entries[peer_index as usize].peer_asn,
            peer_ip: peer_index_table.entries[peer_index as usize].peer_ip,
            bgp_origin: None,
            bgp_as_path: None,
            bgp_next_hop: None,
            bgp_community: None,
            bgp_large_community: None,
            bgp_multi_exit_disc: None,
        };

        // loop over all attributes
        let mut offset: u16 = 0;
        while offset < rib_entry.attribute_length {
            let header = BgpAttributeHeader::from_reader(reader)?;
            offset += header.offset;
            match header.attribute_type {
                BgpAttributeType::Origin => {
                    rib_entry.bgp_origin = BgpOrigin::from_reader(reader).ok();
                }
                BgpAttributeType::AsPath => {
                    rib_entry.bgp_as_path = BgpAsPath::from_reader(reader).ok();
                }
                BgpAttributeType::NextHop => {
                    rib_entry.bgp_next_hop = BgpNextHop::from_reader(reader).ok();
                }
                BgpAttributeType::Community => {
                    rib_entry.bgp_community =
                        BgpCommunity::from_reader(reader, header.attribute_length).ok();
                }
                BgpAttributeType::LargeCommunity => {
                    rib_entry.bgp_large_community =
                        BgpLargeCommunity::from_reader(reader, header.attribute_length).ok();
                }
                BgpAttributeType::MultiExitDisc => {
                    rib_entry.bgp_multi_exit_disc = BgpMultiExitDisc::from_reader(reader).ok();
                }
                _ => Err(mrt::Error::InvalidBgpAttributeType(header.attribute_type))?,
            }
        }
        println!("{:?}", rib_entry);
        Ok(rib_entry)
    }
}
