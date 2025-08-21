mod mrt;

use chrono::{DateTime, Utc};
use clap::Parser;
use mrt::{
    Error, MRTMessage, MRTSubType, MRTType, Result, message::PeerIndexTable,
    message::RibIpV4Unicast,
};

use std::{
    fs::File,
    io::{BufReader, BufWriter, Cursor, Read, prelude::*},
    path::Path,
    process::exit,
};

/// Simple program to greet a person
#[derive(Parser, Debug)]
#[command(name = "mrtdump")]
#[command(version = env!("CARGO_PKG_VERSION"))]
#[command(about = "Read MRT binary files (*, .gz, .bz2) and format and print it in a human-readable format JSON/CSV/MULTILINE", long_about = None)]
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
    /// Input file path MRT format *.gz, *.bz2 or raw
    input_file: String,
    #[arg(short, long)]
    /// Optional Output file path
    output_file: Option<String>,
}

fn open_file_without_extension(filename: &str, capacity: usize) -> Result<Box<dyn Read>> {
    let file = File::open(filename)?;
    Ok(Box::new(BufReader::with_capacity(capacity, file)))
}

fn open_file_gz(filename: &str) -> Result<Box<dyn Read>> {
    let file = File::open(filename)?;
    let decoder = flate2::read::GzDecoder::new(file);
    Ok(Box::new(decoder))
}

fn open_file_bz2(filename: &str) -> Result<Box<dyn Read>> {
    let bz2 = File::open(filename)?;
    let decoder = bzip2::read::BzDecoder::new(bz2);
    Ok(Box::new(decoder))
}

fn open_file(filename: &str) -> Result<Box<dyn Read>> {
    const BUFFER_SIZE: usize = 1024 * 1024; // 1 MB buffer size
    let file_ext = Path::new(&filename).extension();
    if file_ext.is_none() {
        return open_file_without_extension(filename, BUFFER_SIZE);
    }

    match file_ext.unwrap().to_str() {
        Some("gz") => open_file_gz(filename),
        Some("bz2") => open_file_bz2(filename),
        _ => open_file_without_extension(filename, BUFFER_SIZE),
    }
}

fn gen_writer(file: &Option<String>) -> Result<Box<dyn std::io::Write>> {
    match file {
        Some(path) => {
            let file = File::create(path)?;
            Ok(Box::new(BufWriter::new(file)))
        }
        None => Ok(Box::new(BufWriter::new(std::io::stdout()))),
    }
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
        rib_ipv4_unicast.write_json_records(writer)?;
    } else if arg.csv {
        rib_ipv4_unicast.write_csv_records(writer)?;
    } else {
        rib_ipv4_unicast.write_multiline_records(writer)?;
    }
    Ok(())
}

fn read_table_dump_v2<R: Read, W: Write>(
    reader: &mut R,
    writer: &mut W,
    peer_index_table: &mut Cursor<Vec<u8>>,
    arg: &Args,
) -> Result<()> {
    // Read the table dump v2
    let peer_index_table = PeerIndexTable::from_reader(peer_index_table)?;
    while let Ok(mut message) = MRTMessage::from_reader(reader) {
        // Match the message type and subtype
        match (message.header.mrt_type, message.header.mrt_subtype) {
            (MRTType::TableDumpV2, MRTSubType::RibIpV4Unicast) => read_rib_ipv4_unicast(
                &mut message.payload,
                writer,
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
    writer.flush()?;
    Ok(())
}

fn main() -> Result<()> {
    let args = Args::parse();
    // open the file
    let mut file = open_file(&args.input_file).unwrap_or_else(|_| {
        eprintln!("Failed to open file: {}", args.input_file);
        exit(1);
    });

    let mut writer = gen_writer(&args.output_file)?;

    // Read the first message
    let mut message = MRTMessage::from_reader(&mut file)?;

    // Match the first message type and subtype and parse it
    match (message.header.mrt_type, message.header.mrt_subtype) {
        (MRTType::TableDumpV2, MRTSubType::PeerIndexTable) => {
            // Read the peer index table and print the subsequent messages associated to it
            read_table_dump_v2(&mut file, &mut writer, &mut message.payload, &args).unwrap_or_else(
                |e| {
                    eprintln!("Error reading table dump v2, {}", e);
                    exit(1);
                },
            )
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
