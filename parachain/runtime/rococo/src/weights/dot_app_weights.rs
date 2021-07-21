
//! Autogenerated weights for dot_app
//!
//! THIS FILE WAS AUTO-GENERATED USING THE SUBSTRATE BENCHMARK CLI VERSION 3.0.0
//! DATE: 2021-05-08, STEPS: `[50, ]`, REPEAT: 20, LOW RANGE: `[]`, HIGH RANGE: `[]`
//! EXECUTION: Some(Wasm), WASM-EXECUTION: Compiled, CHAIN: Some("/tmp/snowbridge-benchmark-bNy/spec.json"), DB CACHE: 128

// Executed Command:
// target/release/snowbridge
// benchmark
// --chain
// /tmp/snowbridge-benchmark-bNy/spec.json
// --execution
// wasm
// --wasm-execution
// compiled
// --pallet
// dot_app
// --extrinsic
// *
// --repeat
// 20
// --steps
// 50
// --output
// runtime/rococo/src/weights/dot_app_weights.rs


#![allow(unused_parens)]
#![allow(unused_imports)]

use frame_support::{traits::Get, weights::Weight};
use sp_std::marker::PhantomData;

/// Weight functions for dot_app.
pub struct WeightInfo<T>(PhantomData<T>);
impl<T: frame_system::Config> dot_app::WeightInfo for WeightInfo<T> {
	fn lock() -> Weight {
		(168_259_000 as Weight)
			.saturating_add(T::DbWeight::get().reads(4 as Weight))
			.saturating_add(T::DbWeight::get().writes(3 as Weight))
	}
	fn unlock() -> Weight {
		(101_556_000 as Weight)
			.saturating_add(T::DbWeight::get().reads(3 as Weight))
			.saturating_add(T::DbWeight::get().writes(2 as Weight))
	}
}
