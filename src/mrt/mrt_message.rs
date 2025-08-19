use super::{Error, MRTHeader};
use std::io::{Cursor, Read};

#[derive(Debug)]
pub struct MRTMessage {
    pub header: MRTHeader,
    pub payload: Cursor<Vec<u8>>,
}

impl MRTMessage {
    pub fn from_reader<R: Read>(reader: &mut R) -> Result<Self, Error> {
        let header = MRTHeader::from_reader(reader)?;
        let mut payload = vec![0u8; header.length as usize];
        reader.read_exact(&mut payload)?;
        Ok(MRTMessage {
            header,
            payload: Cursor::new(payload),
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Cursor;

    #[test]
    fn test_mrt_message_parsing() {
        let mut cursor = Cursor::new(vec![
            0, 0, 0, 0, // ts
            0, 0x0d, // mrt_type
            0, 0x01, // mrt_subtype
            0, 0, 0, 0x04, // length
            0x1, 0x1, 0x1, 0x1, // payload
        ]);
        let message = MRTMessage::from_reader(&mut cursor);
        assert!(message.is_ok());
        let mut buf = [0u8; 4];
        message.unwrap().payload.read_exact(&mut buf).unwrap();
        assert_eq!(buf, [0x1, 0x1, 0x1, 0x1]);
    }
}
