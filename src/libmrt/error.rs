use thiserror::Error;

#[derive(Debug, Error)]
pub enum LibMrtError {
    #[error("Bad MRT type or unsupported MRT type")]
    BadMrtType,
    #[error("Bad MRT subtype or unsupported MRT subtype")]
    BadMrtSubtype,
    #[error("Unable to parse MRT header")]
    BadMrtHeader,
}
