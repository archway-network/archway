use serde::de::{DeserializeOwned};
use std::collections::HashMap;

use cosmwasm_std::testing::{BankQuerier, MockApi, MockStorage ,MockQuerierCustomHandlerResult, MOCK_CONTRACT_ADDR };
#[cfg(feature = "staking")]
use cosmwasm_std::testing::StakingQuerier;
use cosmwasm_std::{
  from_slice, Coin, CustomQuery, Empty, Querier, QuerierResult, QueryRequest,
  SystemError, SystemResult, WasmQuery, OwnedDeps 
};

pub fn mock_dependencies(
    contract_balance: &[Coin],
) -> OwnedDeps<MockStorage, MockApi, MockQuerier> {
    OwnedDeps {
        storage: MockStorage::default(),
        api: MockApi::default(),
        querier: MockQuerier::new(&[(MOCK_CONTRACT_ADDR, contract_balance)]),
    }
}

/// https://docs.rs/cosmwasm-std/0.16.1/src/cosmwasm_std/mock.rs.html#384-395
/// adds a custom wasm handler to be injected to the querier


/// MockQuerier holds an immutable table of bank balances
pub struct MockQuerier<C: DeserializeOwned = Empty> {
    bank: BankQuerier,
    #[cfg(feature = "staking")]
    staking: StakingQuerier,
    // placeholder to add support later
    wasm: WasmQuerier,

    /// A handler to handle custom queries. This is set to a dummy handler that
    /// always errors by default. Update it via `with_custom_handler`.
    ///Use box to avoid the need of another generic type

    custom_handler: Box<dyn for<'a> Fn(&'a C) -> MockQuerierCustomHandlerResult>,
}

impl<C: DeserializeOwned> MockQuerier<C> {
    pub fn new(balances: &[(&str, &[Coin])]) -> Self {
        MockQuerier {
            bank: BankQuerier::new(balances),
            #[cfg(feature = "staking")]
            staking: StakingQuerier::default(),
            wasm: WasmQuerier::new(balances),
            // strange argument notation suggested as a workaround here: https://github.com/rust-lang/rust/issues/41078#issuecomment-294296365
            custom_handler: Box::from(|_: &_| -> MockQuerierCustomHandlerResult {
                SystemResult::Err(SystemError::UnsupportedRequest {
                    kind: "custom".to_string(),
                })
            }),
        }
    }

    #[cfg(feature = "staking")]
    pub fn update_staking(
        &mut self,
        denom: &str,
        validators: &[crate::query::Validator],
        delegations: &[crate::query::FullDelegation],
    ) {
        self.staking = StakingQuerier::new(denom, validators, delegations);
    }

    pub fn with_custom_handler<CH: 'static>(mut self, handler: CH) -> Self
    where
        CH: Fn(&C) -> MockQuerierCustomHandlerResult,
    {
        self.custom_handler = Box::from(handler);
        self
    }

    pub fn with_custom_wasm_contract<CH: 'static>(mut self, handler: CH) -> Self
    where
        CH: Fn(&WasmQuery) -> MockQuerierCustomHandlerResult,
    {
        self.wasm.custom_wasm_handler = Box::from(handler);
        self
    }
}

impl<C: CustomQuery + DeserializeOwned> Querier for MockQuerier<C> {
    fn raw_query(&self, bin_request: &[u8]) -> QuerierResult {
        let request: QueryRequest<C> = match from_slice(bin_request) {
            Ok(v) => v,
            Err(e) => {
                return SystemResult::Err(SystemError::InvalidRequest {
                    error: format!("Parsing query request: {}", e),
                    request: bin_request.into(),
                })
            }
        };
        self.handle_query(&request)
    }
}

impl<C: CustomQuery + DeserializeOwned> MockQuerier<C> {
    pub fn handle_query(&self, request: &QueryRequest<C>) -> QuerierResult {
        match &request {
            QueryRequest::Bank(bank_query) => self.bank.query(bank_query),
            QueryRequest::Custom(custom_query) => (*self.custom_handler)(custom_query),
            #[cfg(feature = "staking")]
            QueryRequest::Staking(staking_query) => self.staking.query(staking_query),
            QueryRequest::Wasm(msg) => self.wasm.query(msg),
            #[cfg(feature = "stargate")]
            QueryRequest::Stargate { .. } => SystemResult::Err(SystemError::UnsupportedRequest {
                kind: "Stargate".to_string(),
            }),
            #[cfg(feature = "stargate")]
            QueryRequest::Ibc(_) => SystemResult::Err(SystemError::UnsupportedRequest {
                kind: "Ibc".to_string(),
            }),
        }
    }
}

struct WasmQuerier {
  contracts: HashMap<String, Vec<Coin>>,
  custom_wasm_handler: Box<dyn for<'a> Fn(&'a WasmQuery) -> MockQuerierCustomHandlerResult>,
}
impl WasmQuerier {
    fn new(balances: &[(&str, &[Coin])]) -> Self {
        let mut map = HashMap::new();
        for (addr, coins) in balances.iter() {
            map.insert(addr.to_string(), coins.to_vec());
        }
        WasmQuerier {
          contracts: map,
          custom_wasm_handler: Box::from(|_: &_| -> MockQuerierCustomHandlerResult {
                SystemResult::Err(SystemError::UnsupportedRequest {
                    kind: "wasm_contract".to_string(),
                })
            })
        }
    }
    fn query(&self, request: &WasmQuery) -> QuerierResult {
        let addr = match request {
            WasmQuery::Smart { contract_addr, .. } => contract_addr,
            WasmQuery::Raw { contract_addr, .. } => contract_addr,
        }
        .clone();
          //SystemResult::Err(SystemError::NoSuchContract { addr })

          (*self.custom_wasm_handler)(request)
    }
}
