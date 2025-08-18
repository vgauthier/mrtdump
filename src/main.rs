mod mrt;

use mrt::{
    Error, MRTMessage, MRTSubType, MRTType, Result, message::PeerIndexTable,
    message::RibIpV4Unicast,
};
use serde_json::to_string_pretty;

use std::{
    fs::File,
    io::{BufReader, Cursor, Read},
    process::exit,
};

fn open_file(path: &str) -> Result<BufReader<File>> {
    let file = File::open(path)?;
    Ok(BufReader::new(file))
}

fn read_table_dump_v2<R: Read>(reader: &mut R, peer_index_table: &Vec<u8>) -> Result<()> {
    // Read the table dump v2
    let mut message_reader = Cursor::new(peer_index_table);
    let peer_index_table = PeerIndexTable::from_reader(&mut message_reader)?;

    // second message
    let message = MRTMessage::from_reader(reader)?;
    let mut message_reader = Cursor::new(message.payload);
    match (message.header.mrt_type, message.header.mrt_subtype) {
        (MRTType::TableDumpV2, MRTSubType::RibIpV4Unicast) => {
            let rib_ipv4_unicast = RibIpV4Unicast::from_reader(
                &mut message_reader,
                &peer_index_table,
                message.header.ts,
            )?;
            println!("{}", rib_ipv4_unicast);
        }
        _ => return Err(mrt::Error::InvalidMrtType),
    }
    Ok(())
}

fn main() -> Result<()> {
    // open the file
    let path = "/Users/vgauthier/Downloads/rib.20250701.0000";
    let mut file = open_file(path).unwrap_or_else(|_| {
        eprintln!("Failed to open file: {}", path);
        exit(1);
    });

    // Read the first header
    let message = MRTMessage::from_reader(&mut file)?;

    // read the message and the following message according to its type and subtype
    match (message.header.mrt_type, message.header.mrt_subtype) {
        (MRTType::TableDumpV2, MRTSubType::PeerIndexTable) => {
            println!("Peer Index Table message found");
            read_table_dump_v2(&mut file, &message.payload)?;
        }
        (t1, t2) => {
            eprintln!("Unexpected message type:< {:?} - {:?}", t1, t2);
            exit(1);
        }
    }
    Ok(())
}
