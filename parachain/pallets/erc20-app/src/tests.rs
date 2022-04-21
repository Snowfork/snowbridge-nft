use crate::mock::{new_tester, Event, System, AccountId, Origin, Assets, ERC20};
use frame_support::{assert_ok};
use sp_keyring::AccountKeyring as Keyring;
use sp_core::H160;
use artemis_core::{ChannelId, AssetId, MultiAsset};

use crate::RawEvent;

fn last_event() -> Event {
	System::events().pop().expect("Event expected").event
}

#[test]
fn mints_after_handling_ethereum_event() {
	new_tester().execute_with(|| {
		let peer_contract = H160::repeat_byte(1);
		let token = H160::repeat_byte(2);
		let sender = H160::repeat_byte(3);
		let recipient: AccountId = Keyring::Bob.into();
		let amount = 10;
		assert_ok!(
			ERC20::mint(
				artemis_dispatch::Origin(peer_contract).into(),
				token,
				sender,
				recipient.clone(),
				amount.into()
			)
		);
		assert_eq!(Assets::balance(AssetId::Token(token), &recipient), amount.into());

		assert_eq!(
			Event::erc20_app(RawEvent::Minted(token, sender, recipient, amount.into())),
			last_event()
		);
	});
}

#[test]
fn burn_should_emit_bridge_event() {
	new_tester().execute_with(|| {
		let token_id = H160::repeat_byte(1);
		let recipient = H160::repeat_byte(2);
		let bob: AccountId = Keyring::Bob.into();
		Assets::deposit(AssetId::Token(token_id), &bob, 500.into()).unwrap();

		assert_ok!(ERC20::burn(
			Origin::signed(bob.clone()),
			ChannelId::Incentivized,
			token_id,
			recipient.clone(),
			20.into()));

		assert_eq!(
			Event::erc20_app(RawEvent::Burned(token_id, bob, recipient, 20.into())),
			last_event()
		);
	});
}
