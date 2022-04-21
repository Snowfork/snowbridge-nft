use frame_support::{
    parameter_types,
    weights::{RuntimeDbWeight, Weight},
    weights::constants::{WEIGHT_PER_MICROS, WEIGHT_PER_MILLIS},
};

parameter_types! {
    /// Weight of importing a block with 0 txs
    pub const BlockExecutionWeight: Weight = 9 * WEIGHT_PER_MILLIS;
    /// Weight of executing 10,000 System remarks (no-op) txs
    pub const ExtrinsicBaseWeight: Weight = 297 * WEIGHT_PER_MICROS;
    /// Weight of reads and writes to RocksDB, the default DB used by Substrate
    pub const RocksDbWeight: RuntimeDbWeight = RuntimeDbWeight {
        read: 30 * WEIGHT_PER_MICROS,
        write: 112 * WEIGHT_PER_MICROS,
    };
}
