//! # ETH
//!
//! An application that implements a bridged ETH asset.
//!
//! ## Overview
//!
//! ETH balances are stored in the tightly-coupled [`asset`] runtime module. When an account holder burns
//! some of their balance, a `Transfer` event is emitted. An external relayer will listen for this event
//! and relay it to the other chain.
//!
//! ## Interface
//!
//! This application implements the [`Application`] trait and conforms to its interface
//!
//! ### Dispatchable Calls
//!
//! - `burn`: Burn an ETH balance.
//!
#![cfg_attr(not(feature = "std"), no_std)]

use frame_system::{self as system, ensure_signed};
use frame_support::{
	decl_error, decl_event, decl_module, decl_storage,
	dispatch::DispatchResult,
};
use sp_std::prelude::*;
use sp_std::convert::TryInto;
use sp_core::{H160, U256};

use artemis_core::{Application, MessageCommitment, SingleAsset};
use artemis_ethereum::Log;

mod payload;
use payload::{InboundPayload, OutboundPayload};

#[cfg(test)]
mod mock;

#[cfg(test)]
mod tests;

pub trait Config: system::Config {
	type Event: From<Event<Self>> + Into<<Self as system::Config>::Event>;

	type Asset: SingleAsset<<Self as system::Config>::AccountId>;

	type MessageCommitment: MessageCommitment;
}

decl_storage! {
	trait Store for Module<T: Config> as EthModule {
		Address get(fn address) config(): H160;
	}
}

decl_event!(
    /// Events for the ETH module.
	pub enum Event<T>
	where
		AccountId = <T as system::Config>::AccountId
	{
		Burned(AccountId, U256),
		Minted(AccountId, U256),
	}
);

decl_error! {
	pub enum Error for Module<T: Config> {
		/// The submitted payload could not be decoded.
		InvalidPayload,
	}
}

decl_module! {

	pub struct Module<T: Config> for enum Call where origin: T::Origin {

		type Error = Error<T>;

		fn deposit_event() = default;

		// Users should burn their holdings to release funds on the Ethereum side
		// TODO: Calculate weights
		#[weight = 0]
		pub fn burn(origin, recipient: H160, amount: U256) -> DispatchResult {
			let who = ensure_signed(origin)?;

			T::Asset::withdraw(&who, amount)?;

			let message = OutboundPayload {
				sender_addr: who.clone(),
				recipient_addr: recipient,
				amount: amount
			};
			T::MessageCommitment::add(Self::address(), 0, &message.encode());
			Self::deposit_event(RawEvent::Burned(who.clone(), amount));

			Ok(())
		}

	}
}

impl<T: Config> Module<T> {

	fn handle_event(payload: InboundPayload<T::AccountId>) -> DispatchResult {
		T::Asset::deposit(&payload.recipient_addr, payload.amount)?;
		Self::deposit_event(RawEvent::Minted(payload.recipient_addr.clone(), payload.amount));
		Ok(())
	}
}

impl<T: Config> Application for Module<T> {
	fn handle(payload: &[u8]) -> DispatchResult {
		// Decode ethereum Log event from RLP-encoded data, and try to convert to InboundPayload
		let payload_decoded = rlp::decode::<Log>(payload)
			.map_err(|_| Error::<T>::InvalidPayload)?
			.try_into()
			.map_err(|_| Error::<T>::InvalidPayload)?;

		Self::handle_event(payload_decoded)
	}

	fn address() -> H160 {
		Address::get()
	}
}
