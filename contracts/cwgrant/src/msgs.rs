use cosmwasm_schema::{cw_serde};
use cosmwasm_std::{Binary, Coin};

#[cw_serde]
pub struct InstantiateMsg {
    pub grants: Vec<String>
}

#[cw_serde]
pub enum SudoMsg {
    CwGrant(CwGrant)
}
#[cw_serde]
pub struct CwGrant {
    pub fee_requested: Vec<Coin>,
    pub msgs: Vec<CwGrantMessage>,
}
#[cw_serde]
pub struct CwGrantMessage {
    pub sender: String,
    pub type_url: String,
    pub msg: Binary,
}

#[cfg(test)]
mod test {
    #[test]
    fn build() {}
}