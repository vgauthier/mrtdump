use super::{
    BgpAsPath, BgpAttributeHeader, BgpAttributeType, BgpCommunity, BgpLargeCommunity,
    BgpMultiExitDisc, BgpNextHop, BgpOrigin,
};
use crate::libmrt::LibMrtError;
use anyhow::Result;
use byteorder::{BigEndian, ReadBytesExt};
use std::default::Default;
use std::io::Read;

#[derive(Debug, Default)]
#[allow(dead_code)]
pub struct RibEntry {
    pub peer_index: u16,
    pub originated_time: u32,
    pub attribute_length: u16,
    pub bgp_origin: Option<BgpOrigin>,
    pub bgp_as_path: Option<BgpAsPath>,
    pub bgp_next_hop: Option<BgpNextHop>,
    pub bgp_community: Option<BgpCommunity>,
    pub bgp_large_community: Option<BgpLargeCommunity>,
    pub bgp_multi_exit_disc: Option<BgpMultiExitDisc>,
}

impl RibEntry {
    pub fn from_reader<R: Read>(reader: &mut R) -> Result<Self> {
        let peer_index = reader.read_u16::<BigEndian>()?;
        let originated_time = reader.read_u32::<BigEndian>()?;
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
            ..Default::default()
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
                _ => Err(LibMrtError::InvalidBgpAttributeType(header.attribute_type))?,
            }
        }
        println!("{:?}", rib_entry);
        Ok(rib_entry)
    }
}
