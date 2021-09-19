#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{Addr, to_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult, QueryRequest, WasmQuery};
use noop_counter::msg::{ QueryMsg as NoopQueryMsg, CountResponse as NoopCountResponse };


use crate::error::ContractError;
use crate::msg::{CountResponse, ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::{State, STATE};

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
        counter_contract:  Addr::unchecked(msg.counter_contract.clone()),
    };
    STATE.save(deps.storage, &state)?;

    Ok(Response::new()
        .add_attribute("method", "instantiate")
        .add_attribute("owner", info.sender)
        .add_attribute("count", msg.count.to_string()))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::Increment {} => try_increment(deps),
        ExecuteMsg::Reset { count } => try_reset(deps, info, count),
    }
}

pub fn try_increment(deps: DepsMut) -> Result<Response, ContractError> {
    if let Ok(contract_counter) = query_counter(&deps) {
        STATE.update(deps.storage, |mut state| -> Result<_, ContractError> {
            state.count += contract_counter.count; // state.count will add n comming from noop contract
            Ok(state)
        })?;
    }
    Ok(Response::new().add_attribute("method", "try_increment"))
}

pub fn query_counter(deps: &DepsMut) -> Result<NoopCountResponse, ContractError> {
    let state = STATE.load(deps.storage)?;

    let msg = NoopQueryMsg::GetCount {};
    let req = QueryRequest::Wasm(WasmQuery::Smart {
        contract_addr: state.counter_contract.into(),
        msg: to_binary(&msg)?,
    });

     match deps.querier.query(&req) {
         Ok(c) => Ok(c),
         Err(e) => Err(ContractError::from(e)),
     }
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
    use cosmwasm_std::testing::{mock_env, mock_info,MOCK_CONTRACT_ADDR};
    use cosmwasm_std::{coins,ContractResult, from_binary, to_binary, Addr, Empty};
    use cw_multi_test::{App, AppBuilder, Contract, ContractWrapper, Executor};
    use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
    use arch_mocks::querier::mock_dependencies_with_wasm_query;

    fn custom_wasm_execute(query: &WasmQuery) -> ContractResult<Binary> {
      let count: i32 = match query {
        WasmQuery::Smart { .. } => 1,
        WasmQuery::Raw { .. } => 0,
        _ => 0
      };
      to_binary(&NoopCountResponse{ count }).into()
    }
    fn mock_app () -> App {
        AppBuilder::new().build()
    }

    fn query_state_contract() -> Box<dyn Contract<Empty>>{
        let contract = ContractWrapper::new(
            crate::contract::execute,
            crate::contract::instantiate,
            crate::contract::query,
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



    static CONTRACT: fn(&WasmQuery) -> ContractResult<Binary> = custom_wasm_execute;
    #[test]
    fn proper_initialization() {
        let mut deps = mock_dependencies_with_wasm_query(&[], &CONTRACT);

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
        let mut deps = mock_dependencies_with_wasm_query(&coins(2, "token"), &CONTRACT);

        let msg = InstantiateMsg { count: 17, counter_contract: String::from(MOCK_CONTRACT_ADDR) };
        let info = mock_info("creator", &coins(2, "token"));
        let _res = instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();

        // beneficiary can release it
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
        let mut deps = mock_dependencies_with_wasm_query(&coins(2, "token"), &CONTRACT);

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
    fn query_external_contract_integration() {
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
        let query_state_msg = InstantiateMsg {
            count: 0,
            counter_contract: noop_addr.clone().into_string(),
        };
        let query_state_addr = router.instantiate_contract(
                query_state_id, owner.clone(), &query_state_msg,
                &[], "query_state", None,
            ).unwrap();

        assert_ne!(noop_addr.clone(), query_state_addr.clone());
        let query_state_execute_msg = ExecuteMsg::Increment {};
        let res = router.execute_contract(owner.clone(), query_state_addr.clone(), &query_state_execute_msg,&[]).unwrap();
        println!("{:?}", res.events);
        assert_eq!(2, res.events.len());

        let count_res: CountResponse = router.wrap().query_wasm_smart(&query_state_addr, &QueryMsg::GetCount{}).unwrap();
        println!("{:?}", count_res);
        assert_eq!(2, count_res.count);
    }
}
