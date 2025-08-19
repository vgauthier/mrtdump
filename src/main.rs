mod mrt;

use chrono::{DateTime, Utc};
use clap::Parser;
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

/// Simple program to greet a person
#[derive(Parser, Debug)]
#[command(name = "mrtdump")]
#[command(version = "0.1")]
#[command(about = "Read MRT binary files and format and print it in a human-readable format JSON/CSV/PRINT", long_about = None)]
struct Args {
    #[arg(short, long, default_value_t = false)]
    /// Output in JSON format
    json: bool,
    /// Input file path MRT format
    input_file: String,
}

fn open_file(path: &str) -> Result<BufReader<File>> {
    let file = File::open(path)?;
    Ok(BufReader::new(file))
}

fn read_rib_ipv4_unicast<R: Read>(
    reader: &mut R,
    peer_index_table: &PeerIndexTable,
    timestamp: DateTime<Utc>,
    arg: &Args,
) -> Result<()> {
    let rib_ipv4_unicast = RibIpV4Unicast::from_reader(reader, peer_index_table, timestamp)?;
    match arg.json {
        true => println!("{}", to_string_pretty(&rib_ipv4_unicast)?),
        false => println!("{}", rib_ipv4_unicast),
    }
    Ok(())
}

fn read_table_dump_v2<R: Read>(
    reader: &mut R,
    peer_index_table: &mut Cursor<Vec<u8>>,
    arg: &Args,
) -> Result<()> {
    // Read the table dump v2
    let peer_index_table = PeerIndexTable::from_reader(peer_index_table)?;

    loop {
        // second message
        let mut message = MRTMessage::from_reader(reader);
        if let Err(e) = message {
            eprintln!("Error reading MRT message: {}", e);
            break;
        }
        let mut message = message.unwrap();
        //let mut message_reader = Cursor::new(message.payload);
        match (message.header.mrt_type, message.header.mrt_subtype) {
            (MRTType::TableDumpV2, MRTSubType::RibIpV4Unicast) => read_rib_ipv4_unicast(
                &mut message.payload,
                &peer_index_table,
                message.header.ts,
                arg,
            )?,
            _ => {
                return Err(Error::InvalidMrtType(
                    message.header.mrt_type,
                    message.header.mrt_subtype,
                ));
            }
        }
    }
    Ok(())
}

fn main() -> Result<()> {
    let args = Args::parse();
    // open the file
    let mut file = open_file(&args.input_file).unwrap_or_else(|_| {
        eprintln!("Failed to open file: {}", args.input_file);
        exit(1);
    });

    // Read the first message
    let mut message = MRTMessage::from_reader(&mut file)?;

    // Match the first message type and subtype and parse it
    match (message.header.mrt_type, message.header.mrt_subtype) {
        (MRTType::TableDumpV2, MRTSubType::PeerIndexTable) => {
            // Read the peer index table and print the subsequent messages associated to it
            read_table_dump_v2(&mut file, &mut message.payload, &args).unwrap_or_else(|e| {
                eprintln!("Error reading table dump v2, {}", e);
                exit(1);
            })
        }
        (t1, t2) => {
            eprintln!(
                "Unable to read entry of type: {:?} with subtype {:?}",
                t1, t2
            );
            exit(1);
        }
    }
    Ok(())
}
