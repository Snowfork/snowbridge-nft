#![cfg_attr(not(feature = "std"), no_std)]

use sp_std::prelude::*;
use frame_support::{
	decl_module, decl_storage, decl_event, decl_error,
	weights::Weight,
	traits::Get,
};

use sp_io::offchain_index;
use sp_core::{H160, RuntimeDebug};
use sp_runtime::{
	traits::{Member, Hash, Zero, MaybeSerializeDeserialize},
	DigestItem
};

use codec::{Encode, Decode};
use artemis_core::MessageCommitment;

use ethabi::{self, Token};

#[cfg(test)]
mod mock;

#[cfg(test)]
mod tests;

/// Custom DigestItem for header digest
#[derive(Encode, Decode, Copy, Clone, PartialEq, RuntimeDebug)]
enum AuxiliaryDigestItem<H: Encode> {
	CommitmentHash(H)
}

impl<T, H: Encode> Into<DigestItem<T>> for AuxiliaryDigestItem<H> {
    fn into(self) -> DigestItem<T> {
        DigestItem::Other(self.encode())
    }
}

#[derive(Encode, Decode, Clone, PartialEq, RuntimeDebug)]
struct Message {
	address: H160,
	nonce: u64,
	payload: Vec<u8>,
}

pub trait Config: frame_system::Config {

	const INDEXING_PREFIX: &'static [u8];

	type Hashing: Hash<Output = <Self as Config>::Hash>;

	type Hash: Member + MaybeSerializeDeserialize + sp_std::fmt::Debug
		+ sp_std::hash::Hash + AsRef<[u8]> + AsMut<[u8]> + Copy + Default + codec::Codec
		+ codec::EncodeLike;

	type Event: From<Event> + Into<<Self as frame_system::Config>::Event>;

	type CommitInterval: Get<Self::BlockNumber>;
}

decl_storage! {
	trait Store for Module<T: Config> as Commitments {
		/// Messages waiting to be committed
		pub MessageQueue: Vec<Message>;
	}
}

decl_event! {
	pub enum Event {

	}
}

decl_error! {
	pub enum Error for Module<T: Config> {}
}

decl_module! {
	pub struct Module<T: Config> for enum Call where origin: T::Origin {
		type Error = Error<T>;

		fn deposit_event() = default;

		// Generate a message commitment every `T::CommitInterval` blocks.
		//
		// The hash of the commitment is stored as a digest item `CustomDigestItem::Commitment`
		// in the block header. The committed messages are persisted into storage.
		fn on_initialize(now: T::BlockNumber) -> Weight {
			if (now % T::CommitInterval::get()).is_zero() {
				Self::commit()
			} else {
				0
			}
		}
	}
}

impl<T: Config> Module<T> {

	fn offchain_key(hash: <T as Config>::Hash) -> Vec<u8> {
		(T::INDEXING_PREFIX, hash).encode()
	}

	// TODO: return proper weight
	fn commit() -> Weight {
		let messages: Vec<Message> = <Self as Store>::MessageQueue::take();
		if messages.len() == 0 {
			return 0
		}

		let commitment = Self::encode_commitment(&messages);
		let commitment_hash = <T as Config>::Hashing::hash(&commitment);

		let digest_item = AuxiliaryDigestItem::CommitmentHash(commitment_hash.clone()).into();
		<frame_system::Module<T>>::deposit_log(digest_item);

		offchain_index::set(&Self::offchain_key(commitment_hash), &commitment);

		0
	}

	fn encode_commitment(commitment: &[Message]) -> Vec<u8> {
		let messages: Vec<Token> = commitment.iter()
			.map(|message|
				Token::Tuple(vec![
					Token::Address(message.address),
					Token::Bytes(message.payload.clone()),
					Token::Uint(message.nonce.into())
				])
			)
			.collect();

		ethabi::encode(&vec![Token::FixedArray(messages)])
	}

}

impl<T: Config> MessageCommitment for Module<T> {

	// Add a message for eventual inclusion in a commitment
	// TODO: Number of messages per commitment should be bounded
	fn add(address: H160, nonce: u64, payload: &[u8]) {
		<Self as Store>::MessageQueue::append(Message { address, nonce, payload: payload.to_vec() });
	}
}
