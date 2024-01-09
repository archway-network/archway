use cosmwasm_std::{Addr, Empty};
use cw_storage_plus::Map;

pub const GRANTS: Map<&Addr, Empty> = Map::new("grants");