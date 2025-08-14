use crate::mrt::message;

#[derive(Debug, thiserror::Error)]
pub enum Error {
    #[error("Bad MRT type or unsupported MRT type")]
    BadMrtType,
    #[error("Bad MRT subtype or unsupported MRT subtype")]
    BadMrtSubtype,
    #[error("Unable to parse MRT header")]
    BadMrtHeader,
    #[error("Invalid BGP attribute type {0}")]
    InvalidBgpAttributeType(message::BgpAttributeType),
}
