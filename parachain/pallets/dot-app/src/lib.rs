#![cfg_attr(not(feature = "std"), no_std)]
use frame_system::{self as system, ensure_signed};
use frame_support::{
	decl_error, decl_event, decl_module, decl_storage,
	dispatch::{DispatchError, DispatchResult},
	transactional,
	traits::{
		Get,
		EnsureOrigin,
		Currency,
		ExistenceRequirement::{KeepAlive, AllowDeath},
	},
	weights::Weight,
};
use sp_std::{
	prelude::*,
};
use sp_core::{H160, U256};
use sp_runtime::{
	ModuleId,
	traits::{StaticLookup, AccountIdConversion},
};

use artemis_core::{ChannelId, OutboundRouter};

use primitives::{wrap, unwrap};

use payload::OutboundPayload;

mod benchmarking;
mod payload;
pub mod primitives;

#[cfg(test)]
mod mock;

#[cfg(test)]
mod tests;

/// Weight functions needed for this pallet.
pub trait WeightInfo {
	fn lock() -> Weight;
	fn unlock() -> Weight;
}

impl WeightInfo for () {
	fn lock() -> Weight { 0 }
	fn unlock() -> Weight { 0 }
}

type BalanceOf<T> = <<T as Config>::Currency as Currency<<T as frame_system::Config>::AccountId>>::Balance;

pub trait Config: system::Config {
	type Event: From<Event<Self>> + Into<<Self as system::Config>::Event>;

	type Currency: Currency<Self::AccountId>;

	type OutboundRouter: OutboundRouter<Self::AccountId>;

	type CallOrigin: EnsureOrigin<Self::Origin, Success=H160>;

	type ModuleId: Get<ModuleId>;

	type Decimals: Get<u32>;
	/// Weight information for extrinsics in this pallet
	type WeightInfo: WeightInfo;
}

decl_storage! {
	trait Store for Module<T: Config> as DotModule {
		/// Address of the peer application on the Ethereum side.
		Address get(fn address) config(): H160;
	}
}

decl_event!(
    /// Events for the ETH module.
	pub enum Event<T>
	where
		AccountId = <T as system::Config>::AccountId,
		Balance = BalanceOf<T>
	{
		Locked(AccountId, H160, Balance),
		Unlocked(H160, AccountId, Balance),
	}
);

decl_error! {
	pub enum Error for Module<T: Config> {
		/// Illegal conversion between native and wrapped DOT.
		///
		/// In practice, this error should never occur under the conditions
		/// we've tested. If however the bridge or the peer Ethereum contract
		/// is exploited, then all bets are off.
		Overflow
	}
}

decl_module! {
	pub struct Module<T: Config> for enum Call where origin: T::Origin {

		type Error = Error<T>;

		fn deposit_event() = default;

		// Verify that `T::Decimals` is 10 (DOT), or 12 (KSM) to guarantee
		// safe conversions between native and wrapped DOT.
		fn integrity_test() {
			sp_io::TestExternalities::new_empty().execute_with(|| {
				let allowed_decimals: &[u32] = &[10, 12];
				let decimals = T::Decimals::get();
				assert!(
					allowed_decimals.contains(&decimals)
				)
			});
		}

		#[weight = T::WeightInfo::lock()]
		#[transactional]
		pub fn lock(origin, channel_id: ChannelId, recipient: H160, amount: BalanceOf<T>) -> DispatchResult {
			let who = ensure_signed(origin)?;

			T::Currency::transfer(&who, &Self::account_id(), amount, AllowDeath)?;

			let amount_wrapped = wrap::<T>(amount, T::Decimals::get()).ok_or(Error::<T>::Overflow)?;

			let message = OutboundPayload {
				sender: who.clone(),
				recipient: recipient.clone(),
				amount: amount_wrapped,
			};

			T::OutboundRouter::submit(channel_id, &who, Address::get(), &message.encode())?;
			Self::deposit_event(RawEvent::Locked(who.clone(), recipient, amount));
			Ok(())
		}

		#[weight = T::WeightInfo::unlock()]
		#[transactional]
		pub fn unlock(origin, sender: H160, recipient: <T::Lookup as StaticLookup>::Source, amount: U256) -> DispatchResult {
			let who = T::CallOrigin::ensure_origin(origin)?;
			if who != Address::get() {
				return Err(DispatchError::BadOrigin.into());
			}

			let amount_unwrapped = unwrap::<T>(amount, T::Decimals::get()).ok_or(Error::<T>::Overflow)?;

			let recipient = T::Lookup::lookup(recipient)?;
			T::Currency::transfer(&Self::account_id(), &recipient, amount_unwrapped, KeepAlive)?;
			Self::deposit_event(RawEvent::Unlocked(sender, recipient, amount_unwrapped));
			Ok(())
		}
	}
}

impl<T: Config> Module<T> {
	pub fn account_id() -> T::AccountId {
		T::ModuleId::get().into_account()
	}
}
