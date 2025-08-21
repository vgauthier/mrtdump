use crate::mrt::{MRTSubType, MRTType};
use thiserror::Error;

pub type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, Error)]
pub enum Error {
    #[error("Bad MRT type or unsupported MRT type: {0}")]
    BadMrtType(u16),
    #[error("Bad MRT subtype or unsupported MRT subtype: {0}")]
    BadMrtSubtype(u16),
    #[error("Unable to parse MRT header")]
    BadMrtHeader,
    #[error("Error parsing BGP attribute number: {0}")]
    ParsingBgpAttribute(u8),
    #[error("Wrong MRT type or subtype")]
    InvalidMrtType(MRTType, MRTSubType),
    #[error("I/O error: {0}")]
    IO(#[from] std::io::Error),
    #[error("UTF-8 error: {0}")]
    FromUtf8(#[from] std::string::FromUtf8Error),
    #[error("Serde error: {0}")]
    Serde(#[from] serde_json::Error),
    #[error("Invalid peer index: {0}")]
    InvalidPeerIndex(u16),
    #[error("Invalid community length: {0}")]
    InvalidCommunityLength(u16),
    #[error("Invalid large community length: {0}")]
    InvalidLargeCommunityLength(u16),
    #[error("CSV error: {0}")]
    CsvSerialization(#[from] csv::Error),
    #[error("Bad RIB entry header")]
    BadRibEntryHeader,
}
