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
    io::{BufReader, BufWriter, Cursor, Read, prelude::*},
    process::exit,
};

/// Simple program to greet a person
#[derive(Parser, Debug)]
#[command(name = "mrtdump")]
#[command(version = "0.1")]
#[command(about = "Read MRT binary files and format and print it in a human-readable format JSON/CSV/PRINT", long_about = None)]
struct Args {
    #[arg(short, long, default_value_t = true)]
    /// Multi-line, human-readable (the default)
    print: bool,
    #[arg(short, long, default_value_t = false)]
    /// Output in JSON format
    json: bool,
    /// Output in CSV format
    #[arg(short, long, default_value_t = false)]
    csv: bool,
    /// Input file path MRT format
    input_file: String,
}

fn open_file(path: &str) -> Result<BufReader<File>> {
    const BUFFER_SIZE: usize = 10 * 1024 * 1024; // 10 MB
    let file = File::open(path)?;
    Ok(BufReader::with_capacity(BUFFER_SIZE, file))
}

fn read_rib_ipv4_unicast<R: Read, W: Write>(
    reader: &mut R,
    writer: &mut W,
    peer_index_table: &PeerIndexTable,
    timestamp: DateTime<Utc>,
    arg: &Args,
) -> Result<()> {
    let rib_ipv4_unicast = RibIpV4Unicast::from_reader(reader, peer_index_table, timestamp)?;
    if arg.json {
        writeln!(writer, "{}", to_string_pretty(&rib_ipv4_unicast)?)?;
    } else if arg.csv {
        rib_ipv4_unicast.write_csv_records(writer)?;
    } else {
        writeln!(writer, "{}", rib_ipv4_unicast.to_string())?;
    }
    Ok(())
}

fn read_table_dump_v2<B: BufRead>(
    reader: &mut B,
    peer_index_table: &mut Cursor<Vec<u8>>,
    arg: &Args,
) -> Result<()> {
    // Read the table dump v2
    let peer_index_table = PeerIndexTable::from_reader(peer_index_table)?;
    let mut writer = BufWriter::new(std::io::stdout());
    while let Ok(mut message) = MRTMessage::from_reader(reader) {
        // Match the message type and subtype
        match (message.header.mrt_type, message.header.mrt_subtype) {
            (MRTType::TableDumpV2, MRTSubType::RibIpV4Unicast) => read_rib_ipv4_unicast(
                &mut message.payload,
                &mut writer,
                &peer_index_table,
                message.header.ts,
                arg,
            )
            .unwrap_or_else(|e| {
                eprintln!("Error reading RIB IPv4 Unicast: {} skip the entry", e);
            }),
            _ => {
                return Err(Error::InvalidMrtType(
                    message.header.mrt_type,
                    message.header.mrt_subtype,
                ));
            }
        }
    }
    writer.flush().unwrap();
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
