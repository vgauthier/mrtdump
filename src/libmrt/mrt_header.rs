use super::LibMrtError;
use anyhow::Result;
use byteorder::{BigEndian, ReadBytesExt};
use chrono::{Utc, prelude::DateTime};
use std::io::Read;
use strum_macros::FromRepr;

#[derive(Debug, PartialEq, FromRepr)]
#[repr(u16)]
pub enum MRTSubType {
    PeerIndexTable = 1, // Peer index type
    RibIpv4unicast = 2, // RIB IPv4 Unicast subtype
}

#[derive(Debug, PartialEq, FromRepr)]
#[repr(u16)]
pub enum MRTType {
    OspfV2 = 11,
    TableDump = 12,
    TableDumpV2 = 13,
    Bgp4Mp = 16,
    Bgp4MpEt = 17,
    Isis = 32,
    IsisEt = 33,
    OspfV3 = 48,
    OspfV3Et = 49,
}

#[derive(Debug)]
#[allow(dead_code)]
pub struct MRTHeader {
    pub ts: DateTime<Utc>,       // "Timestamp" in seconds since epoch"
    pub mrt_type: MRTType,       // Type of the MRT header message
    pub mrt_subtype: MRTSubType, // Subtype of the MRT header message
    pub length: u32,             // Length of the MRT header message
}

impl MRTHeader {
    pub fn from_reader<R: Read>(reader: &mut R) -> Result<Self> {
        let ts = reader.read_u32::<BigEndian>()?;
        let mrt_type = reader.read_u16::<BigEndian>()?;
        let mrt_subtype = reader.read_u16::<BigEndian>()?;
        let length = reader.read_u32::<BigEndian>()?;
        let mrt_type = MRTType::from_repr(mrt_type).ok_or(LibMrtError::BadMrtType)?;
        let mrt_subtype = MRTSubType::from_repr(mrt_subtype).ok_or(LibMrtError::BadMrtSubtype)?;
        let ts = DateTime::from_timestamp(ts as i64, 0).ok_or(LibMrtError::BadMrtHeader)?;
        Ok(MRTHeader {
            ts,
            mrt_type,
            mrt_subtype,
            length,
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Cursor;
    #[test]
    fn test_reading_mrt_header() {
        let mut cursor = Cursor::new(vec![
            0, 0, 0, 0, // ts
            0, 0x0d, // mrt_type
            0, 0x01, // mrt_subtype
            0, 0, 0, 0, // length
        ]);
        let header = MRTHeader::from_reader(&mut cursor).unwrap();
        assert_eq!(header.ts, DateTime::from_timestamp(0, 0).unwrap());
        assert_eq!(header.mrt_type, MRTType::TableDumpV2);
        assert_eq!(header.mrt_subtype, MRTSubType::PeerIndexTable);
        assert_eq!(header.length, 0);
    }

    #[test]
    fn test_reading_mrt_header_bad_type() {
        let mut cursor = Cursor::new(vec![
            0, 0, 0, 0, // ts
            0, 0x13, // mrt_type
            0, 0x01, // mrt_subtype
            0, 0, 0, 0, // length
        ]);
        let header = MRTHeader::from_reader(&mut cursor);
        assert!(header.is_err());
        let error_msg = header.unwrap_err().to_string();
        let expected_error = LibMrtError::BadMrtType.to_string();
        assert_eq!(error_msg, expected_error);
    }

    #[test]
    fn test_reading_mrt_header_bad_subtype() {
        let mut cursor = Cursor::new(vec![
            0, 0, 0, 0, // ts
            0, 0x0d, // mrt_type
            0, 0x05, // mrt_subtype
            0, 0, 0, 0, // length
        ]);
        let header = MRTHeader::from_reader(&mut cursor);
        assert!(header.is_err());
        let error_msg = header.unwrap_err().to_string();
        let expected_error = LibMrtError::BadMrtSubtype.to_string();
        assert_eq!(error_msg, expected_error);
    }
}
