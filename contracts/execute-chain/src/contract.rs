#[cfg(not(feature = "library"))]
use std::vec::Vec;

use cosmwasm_std::entry_point;
use cosmwasm_std::{Addr, to_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult, CosmosMsg, WasmMsg};
use query_state::msg::ExecuteMsg as ExternalExecuteMsg;
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::msg::{CountResponse, ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::{State, STATE};

// version info for migration info
const CONTRACT_NAME: &str = "crates.io:execute-chain";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    let state = State {
        count: msg.count,
        owner: info.sender.clone(),
        counter_contract: Addr::unchecked(msg.counter_contract.clone()),
    };
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;
    STATE.save(deps.storage, &state)?;

    Ok(Response::new()
        .add_attribute("method", "instantiate")
        .add_attribute("owner", info.sender)
        .add_attribute("count", msg.count.to_string()))
}

// How to fund
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::Increment {} => try_increment(deps),
        ExecuteMsg::Chain {} => try_chain(deps, _env.contract.address.clone()),
        ExecuteMsg::Reset { count } => try_reset(deps, info, count),
    }
}

pub fn try_chain(deps: DepsMut, contract_address: Addr) -> Result<Response, ContractError> {
    Ok(
        Response::new()
            .add_attribute("method", "try_increment")
            .add_messages(create_msgs(deps, contract_address)?) // TODO: add arguments for msgs
        )
}
fn create_msgs(deps: DepsMut, contract_address: Addr) -> StdResult<Vec<CosmosMsg>>{
    Ok(
        Vec::from([
            create_increment_msg_extern(&deps)?,
            create_increment_msg_intern(contract_address)?,
        ])
    )
}
fn create_increment_msg_extern(deps: &DepsMut) -> StdResult<CosmosMsg> {
    let state = STATE.load(deps.storage)?;
    Ok(CosmosMsg::Wasm(WasmMsg::Execute{
        contract_addr: state.counter_contract.into_string(),
        msg: to_binary(&ExternalExecuteMsg::Increment {})?,
        funds: vec!(),
    }))
}
fn create_increment_msg_intern(contract_address: Addr) -> StdResult<CosmosMsg> {
    Ok(CosmosMsg::Wasm(WasmMsg::Execute{
        contract_addr: contract_address.into_string(), // TODO: owner is not contract_address !
        msg: to_binary(&ExecuteMsg::Increment {})?,
        funds: vec!(),
    }))
}
// TODO: need a helper function that creates an executeMsg from contrat B
pub fn try_increment(deps: DepsMut) -> Result<Response, ContractError> {
    STATE.update(deps.storage, |mut state| -> Result<_, ContractError> {
        state.count += 1;
        Ok(state)
    })?;

    Ok(Response::new().add_attribute("method", "try_increment"))
}
pub fn try_reset(deps: DepsMut, info: MessageInfo, count: i32) -> Result<Response, ContractError> {
    STATE.update(deps.storage, |mut state| -> Result<_, ContractError> {
        if info.sender != state.owner {
            return Err(ContractError::Unauthorized {});
        }
        state.count = count;
        Ok(state)
    })?;
    Ok(Response::new().add_attribute("method", "reset"))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetCount {} => to_binary(&query_count(deps)?),
    }
}

fn query_count(deps: Deps) -> StdResult<CountResponse> {
    let state = STATE.load(deps.storage)?;
    Ok(CountResponse { count: state.count })
}

#[cfg(test)]
mod tests {
    use super::*;
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info, MOCK_CONTRACT_ADDR};
    use cosmwasm_std::{coins, from_binary, Empty};
    use cw_multi_test::{App, AppBuilder, Contract, ContractWrapper, Executor};

    use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};

    fn mock_app () -> App {
        AppBuilder::new().build()
    }

    fn execute_chain_contract() -> Box<dyn Contract<Empty>>{
        let contract = ContractWrapper::new(
            crate::contract::execute,
            crate::contract::instantiate,
            crate::contract::query,
        );
        Box::new(contract)
    }
    fn query_state_contract() -> Box<dyn Contract<Empty>>{
        let contract = ContractWrapper::new(
            query_state::contract::execute,
            query_state::contract::instantiate,
            query_state::contract::query,
        );
        Box::new(contract)
    }

    fn noop_contract() -> Box<dyn Contract<Empty>>{
        let contract = ContractWrapper::new(
            noop_counter::contract::execute,
            noop_counter::contract::instantiate,
            noop_counter::contract::query,
        );
        Box::new(contract)
    }
    #[test]
    fn proper_initialization() {
        let mut deps = mock_dependencies(&[]);

        let msg = InstantiateMsg { count: 17, counter_contract: String::from(MOCK_CONTRACT_ADDR) };
        let info = mock_info("creator", &coins(1000, "earth"));

        // we can just call .unwrap() to assert this was a success
        let res = instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();
        assert_eq!(0, res.messages.len());

        // it worked, let's query the state
        let res = query(deps.as_ref(), mock_env(), QueryMsg::GetCount {}).unwrap();
        let value: CountResponse = from_binary(&res).unwrap();
        assert_eq!(17, value.count);
    }

    #[test]
    fn increment() {
        let mut deps = mock_dependencies(&coins(2, "token"));

        let msg = InstantiateMsg { count: 17, counter_contract: String::from(MOCK_CONTRACT_ADDR) };
        let info = mock_info("creator", &coins(2, "token"));
        let _res = instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();

        // benef_iciary can release it
        let info = mock_info("anyone", &coins(2, "token"));
        let msg = ExecuteMsg::Increment {};
        let _res = execute(deps.as_mut(), mock_env(), info, msg).unwrap();

        // should increase counter by 1
        let res = query(deps.as_ref(), mock_env(), QueryMsg::GetCount {}).unwrap();
        let value: CountResponse = from_binary(&res).unwrap();
        assert_eq!(18, value.count);
    }

    #[test]
    fn reset() {
        let mut deps = mock_dependencies(&coins(2, "token"));

        let msg = InstantiateMsg { count: 17, counter_contract: String::from(MOCK_CONTRACT_ADDR) };
        let info = mock_info("creator", &coins(2, "token"));
        let _res = instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();

        // beneficiary can release it
        let unauth_info = mock_info("anyone", &coins(2, "token"));
        let msg = ExecuteMsg::Reset { count: 5 };
        let res = execute(deps.as_mut(), mock_env(), unauth_info, msg);
        match res {
            Err(ContractError::Unauthorized {}) => {}
            _ => panic!("Must return unauthorized error"),
        }

        // only the original creator can reset the counter
        let auth_info = mock_info("creator", &coins(2, "token"));
        let msg = ExecuteMsg::Reset { count: 5 };
        let _res = execute(deps.as_mut(), mock_env(), auth_info, msg).unwrap();

        // should now be 5
        let res = query(deps.as_ref(), mock_env(), QueryMsg::GetCount {}).unwrap();
        let value: CountResponse = from_binary(&res).unwrap();
        assert_eq!(5, value.count);
    }

    #[test]
    fn chain() {
        let mut deps = mock_dependencies(&coins(2, "token"));

        let msg = InstantiateMsg { count: 17, counter_contract: String::from(MOCK_CONTRACT_ADDR) };
        let info = mock_info("creator", &coins(2, "token"));
        let _res = instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();

        // benef_iciary can release it
        let info = mock_info("anyone", &coins(2, "token"));
        let msg = ExecuteMsg::Chain {};
        let res = execute(deps.as_mut(), mock_env(), info, msg).unwrap();
        assert_eq!(2, res.messages.len());
    }

    #[test]
    fn chain_integration() {
        let mut router = mock_app();

        let owner = Addr::unchecked("owner");
        let init_funds = coins(2000, "arch");
        router.init_bank_balance(&owner, init_funds).unwrap();

        let noop_contract_id = router.store_code(noop_contract());
        let noop_msg = noop_counter::msg::InstantiateMsg {
            count: 2,
        };

        let noop_addr = router.instantiate_contract(
                noop_contract_id, owner.clone(), &noop_msg,
                &[], "Noop", None
            ).unwrap();

        let query_state_id = router.store_code(query_state_contract());
        let query_state_msg = query_state::msg::InstantiateMsg {
            count: 0,
            counter_contract: noop_addr.clone().into_string(),
        };
        let query_state_addr = router.instantiate_contract(
                query_state_id, owner.clone(), &query_state_msg,
                &[], "query_state", None,
            ).unwrap();

        assert_ne!(noop_addr.clone(), query_state_addr.clone());

        let chain_execute_id = router.store_code(execute_chain_contract());
        let chain_instantiate_msg = InstantiateMsg {
            count: 0,
            counter_contract: query_state_addr.clone().into_string(),
        };
        let chain_execute_addr = router.instantiate_contract(
                chain_execute_id, owner.clone(), &chain_instantiate_msg,
                &[], "chain_execute", None
            ).unwrap();

        assert_ne!(query_state_addr.clone(), chain_execute_addr.clone());
        let chain_execute_msg = ExecuteMsg::Chain {};
        let res = match router.execute_contract(owner.clone(), chain_execute_addr.clone(), &chain_execute_msg, &[]) {
            Ok(v) => v,
            Err(e) => {
                println!("{:?}", e);
                panic!("{}",e)
            }
        };
        println!("{:?}", res.events);
        assert_eq!(6, res.events.len());

        let count_res: CountResponse = router.wrap().query_wasm_smart(&chain_execute_addr, &QueryMsg::GetCount{}).unwrap();
        println!("{:?}", count_res);
        assert_eq!(1, count_res.count);

        let count_res: CountResponse = router.wrap().query_wasm_smart(&query_state_addr, &query_state::msg::QueryMsg::GetCount{}).unwrap();
        println!("{:?}", count_res);
        assert_eq!(2, count_res.count);
    }
}
