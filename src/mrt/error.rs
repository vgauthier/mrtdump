use crate::mrt::{MRTSubType, MRTType, message::BgpAttributeType};
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
    #[error("Error parsing BGP attribute")]
    ErrorParsingBgpAttribute,
    #[error("Invalid BGP attribute type {0}")]
    InvalidBgpAttributeType(BgpAttributeType),
    #[error("Wrong MRT type or subtype")]
    InvalidMrtType(MRTType, MRTSubType),
    #[error("I/O error: {0}")]
    IO(#[from] std::io::Error),
    #[error("UTF-8 error: {0}")]
    FromUtf8Error(#[from] std::string::FromUtf8Error),
}
