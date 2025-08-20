use crate::mrt::Error;
use byteorder::{BigEndian, ReadBytesExt};
use serde::Serialize;
use std::{fmt, io::Read, net::Ipv4Addr};
use strum_macros::{Display, FromRepr};

#[derive(Debug, FromRepr, Serialize)]
#[repr(u8)]
pub enum BgpOriginType {
    Igp = 0,
    Egp = 1,
    Incomplete = 2,
}

#[derive(Debug, Serialize)]
#[allow(dead_code)]
pub struct BgpMultiExitDisc(pub u32);

#[derive(Debug, FromRepr, Display, Serialize)]
#[repr(u8)]
pub enum BgpAttributeType {
    Origin = 1,
    AsPath = 2,
    NextHop = 3,
    MultiExitDisc = 4,
    AtomicAggregate = 6,
    Aggregator = 7,
    Community = 8,
    LargeCommunity = 32,
    Otc = 35,
    BfdDiscriminator = 38,
}

#[derive(Debug, Serialize)]
#[allow(dead_code)]
pub struct BgpNextHop(pub Ipv4Addr);

#[derive(Debug, Serialize)]
#[allow(dead_code)]
pub struct BgpOrigin(pub BgpOriginType);

#[derive(Debug, Serialize)]
#[allow(dead_code)]
pub struct BgpCommunity(pub Vec<(u16, u16)>);

#[derive(Debug, Serialize)]
#[allow(dead_code)]
pub struct BgpLargeCommunity(pub Vec<(u32, u32, u32)>);

#[derive(Debug, Serialize)]
#[allow(dead_code)]
pub struct BgpAsPath {
    pub segment_type: u8,
    pub segments: Vec<i32>,
}

#[derive(Debug, Serialize)]
#[allow(dead_code)]
pub struct BgpAggregator {
    pub asn: u32,
    pub ip: Ipv4Addr,
}

#[derive(Debug, Serialize)]
#[allow(dead_code)]
pub struct BgpAttributeHeader {
    pub attribute_flag: u8,
    pub attribute_type: BgpAttributeType,
    pub attribute_length: u16,
    pub offset: u16,
}

impl BgpAttributeHeader {
    pub fn from_reader<R: Read>(reader: &mut R) -> Result<Self, Error> {
        let is_extended_length = 0x10;
        let attribute_flag = reader.read_u8()?;
        let attribute_type = reader.read_u8()?;
        let attribute_length = if attribute_flag & is_extended_length == 0 {
            reader.read_u8()? as u16
        } else {
            reader.read_u16::<BigEndian>()?
        };
        let offset = if attribute_flag & is_extended_length == 0 {
            3 + attribute_length
        } else {
            4 + attribute_length
        };
        // parse bgp attribute
        let attribute_type = BgpAttributeType::from_repr(attribute_type)
            .ok_or(Error::ErrorParsingBgpAttribute(attribute_type))?;

        Ok(BgpAttributeHeader {
            attribute_flag,
            attribute_type,
            attribute_length,
            offset,
        })
    }
}

impl BgpOrigin {
    pub fn from_reader<R: Read>(reader: &mut R) -> Result<Self, Error> {
        let origin = BgpOriginType::from_repr(reader.read_u8()?).ok_or(Error::BadMrtHeader)?;
        Ok(BgpOrigin(origin))
    }
}

impl fmt::Display for BgpOrigin {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self.0 {
            BgpOriginType::Igp => write!(f, "IGP")?,
            BgpOriginType::Egp => write!(f, "EGP")?,
            BgpOriginType::Incomplete => write!(f, "INCOMPLETE")?,
        }
        Ok(())
    }
}

impl BgpAsPath {
    pub fn from_reader<R: Read>(reader: &mut R) -> Result<Self, Error> {
        let segment_type = reader.read_u8()?;
        let segment_count: usize = reader.read_u8()?.into();
        let mut segments = Vec::with_capacity(segment_count);
        for _ in 0..segment_count {
            segments.push(reader.read_i32::<BigEndian>()?);
        }
        Ok(BgpAsPath {
            segment_type,
            segments,
        })
    }
}

impl BgpNextHop {
    pub fn from_reader<R: Read>(reader: &mut R) -> Result<Self, Error> {
        let mut next_hop_bytes = [0u8; 4];
        reader.read_exact(&mut next_hop_bytes)?;
        let ip = Ipv4Addr::from(next_hop_bytes);
        Ok(BgpNextHop(ip))
    }
}

impl BgpAggregator {
    pub fn from_reader<R: Read>(reader: &mut R) -> Result<Self, Error> {
        let asn = reader.read_u32::<BigEndian>()?;
        let mut ip_bytes = [0u8; 4];
        reader.read_exact(&mut ip_bytes)?;
        let ip = Ipv4Addr::from(ip_bytes);
        Ok(BgpAggregator { asn, ip })
    }
}

impl BgpCommunity {
    pub fn from_reader<R: Read>(reader: &mut R, length: u16) -> Result<Self, Error> {
        let community_count: usize = (length / 4).into();
        let mut community = Vec::with_capacity(community_count);
        for _ in 0..community_count {
            let asn = reader.read_u16::<BigEndian>()?;
            let local = reader.read_u16::<BigEndian>()?;
            community.push((asn, local));
        }
        Ok(BgpCommunity(community))
    }
}

impl BgpLargeCommunity {
    pub fn from_reader<R: Read>(reader: &mut R, length: u16) -> Result<Self, Error> {
        let community_count: usize = (length / 12).into();
        let mut community = Vec::with_capacity(community_count);
        for _ in 0..community_count {
            let global_administrator = reader.read_u32::<BigEndian>()?;
            let local_1 = reader.read_u32::<BigEndian>()?;
            let local_2 = reader.read_u32::<BigEndian>()?;
            community.push((global_administrator, local_1, local_2));
        }
        Ok(BgpLargeCommunity(community))
    }
}

impl BgpMultiExitDisc {
    pub fn from_reader<R: Read>(reader: &mut R) -> Result<Self, Error> {
        let metric = reader.read_u32::<BigEndian>()?;
        Ok(BgpMultiExitDisc(metric))
    }
}
