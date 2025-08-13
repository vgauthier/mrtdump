mod libmrt;

use anyhow::Result;
use libmrt::{MRTMessage, message::PeerIndexTable, message::RibIpV4Unicast};
use std::{fs::File, io::Cursor};

fn main() -> Result<()> {
    let path = "/Users/vgauthier/Downloads/rib.20250701.0000";
    let mut file = File::open(path)?;
    // first message suppose to be a PeerIndexTable
    let message = MRTMessage::from_reader(&mut file)?;
    println!("{:?}", message.header);
    let mut message_reader = Cursor::new(message.payload);
    let peer_index_table = PeerIndexTable::from_reader(&mut message_reader)?;
    println!("{:?}", peer_index_table);

    // second message
    let message = MRTMessage::from_reader(&mut file)?;
    let mut message_reader = Cursor::new(message.payload);
    RibIpV4Unicast::from_reader(&mut message_reader)?;
    Ok(())
}
