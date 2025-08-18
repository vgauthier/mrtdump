use crate::mrt::{MRTSubType, MRTType, message};
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
    #[error("Unknown BGP attribute error")]
    UnknownBgpAttribute,
    #[error("Invalid BGP attribute type {0}")]
    InvalidBgpAttributeType(message::BgpAttributeType),
    #[error("Wrong MRT type or subtype")]
    InvalidMrtType,
    #[error("I/O error: {0}")]
    IO(#[from] std::io::Error),
    #[error("UTF-8 error: {0}")]
    FromUtf8Error(#[from] std::string::FromUtf8Error),
}
