mod mrt;

use anyhow::Result;
use mrt::{MRTMessage, message::PeerIndexTable, message::RibIpV4Unicast};
use serde_json::to_string_pretty;
use std::{fs::File, io::BufReader, io::Cursor, process::exit};

fn open_file(path: &str) -> Result<BufReader<File>> {
    let file = File::open(path)?;
    Ok(BufReader::new(file))
}

// fn read_tabledump<R: Read>(reader: &mut R) -> Result<()> {
//     RibEntry::from_reader(reader, peer_index_table)
// }

fn main() -> Result<()> {
    let path = "/Users/vgauthier/Downloads/rib.20250701.0000";
    let mut file = open_file(path).unwrap_or_else(|_| {
        eprintln!("Failed to open file: {}", path);
        exit(1);
    });
    // first message suppose to be a PeerIndexTable
    let message = MRTMessage::from_reader(&mut file)?;
    println!("{:?}", message.header);
    let mut message_reader = Cursor::new(message.payload);
    let peer_index_table = PeerIndexTable::from_reader(&mut message_reader)?;
    println!("{:?}", peer_index_table);

    // second message
    let message = MRTMessage::from_reader(&mut file)?;
    let mut message_reader = Cursor::new(message.payload);
    let rib_ipv4_unicast =
        RibIpV4Unicast::from_reader(&mut message_reader, &peer_index_table, message.header.ts)?;
    println!("{}", rib_ipv4_unicast);
    println!("{}", to_string_pretty(&rib_ipv4_unicast).expect("failled"));
    Ok(())
}
