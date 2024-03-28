#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, GetCountResponse, InstantiateMsg, QueryMsg, SudoMsg};
use crate::state::{State, STATE};

// version info for migration info
const CONTRACT_NAME: &str = "crates.io:callback-test";
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
        error_encountered: false,
    };
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;
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
        ExecuteMsg::Increment {} => execute::increment(deps),
        ExecuteMsg::Reset { count } => execute::reset(deps, info, count),
    }
}

pub mod execute {
    use super::*;

    pub fn increment(deps: DepsMut) -> Result<Response, ContractError> {
        STATE.update(deps.storage, |mut state| -> Result<_, ContractError> {
            state.count += 1;
            Ok(state)
        })?;

        Ok(Response::new().add_attribute("action", "increment"))
    }

    pub fn reset(deps: DepsMut, info: MessageInfo, count: i32) -> Result<Response, ContractError> {
        STATE.update(deps.storage, |mut state| -> Result<_, ContractError> {
            if info.sender != state.owner {
                return Err(ContractError::Unauthorized {});
            }
            state.count = count;
            Ok(state)
        })?;
        Ok(Response::new().add_attribute("action", "reset"))
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetCount {} => to_json_binary(&query::count(deps)?),
    }
}

pub mod query {
    use super::*;

    pub fn count(deps: Deps) -> StdResult<GetCountResponse> {
        let state = STATE.load(deps.storage)?;
        Ok(GetCountResponse { count: state.count, error_encountered: state.error_encountered})
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn sudo(
    deps: DepsMut,
    _env: Env,
    msg: SudoMsg,
) -> Result<Response, ContractError> {
    match msg {
        SudoMsg::Callback { job_id } => sudo::handle_callback(deps, job_id),
        SudoMsg::Error {
            module_name,
            error_code,
            input_payload,
            error_message,
        } => sudo::handle_error(deps, module_name, error_code, input_payload, error_message),
    }
}

pub mod sudo {
    use super::*;
    use std::u64;

    pub fn handle_callback(deps: DepsMut, job_id: u64) -> Result<Response, ContractError> {
        STATE.update(deps.storage, |mut state| -> Result<_, ContractError> {
            if job_id == 0 {
                state.count -= 1;
            };
            if job_id == 1 {
                state.count += 1;
            };
            if job_id == 2 {
                return Err(ContractError::SomeError {});
            }
            Ok(state)
        })?;

        Ok(Response::new().add_attribute("action", "handle_callback"))
    }

    pub fn handle_error(deps: DepsMut, _module_name: String, _error_code: u32, _input_payload: String, _error_message: String) -> Result<Response, ContractError> {
        STATE.update(deps.storage, |mut state| -> Result<_, ContractError> {
            state.error_encountered = true;
            Ok(state)
        })?;
        Ok(Response::new().add_attribute("action", "handle_error"))
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
    use cosmwasm_std::{coins, from_json};
    use tests::sudo::handle_error;

    use crate::contract::sudo::handle_callback;

    #[test]
    fn callback() {
        let mut deps = mock_dependencies();

        let msg = InstantiateMsg { count: 100 };
        let info = mock_info("creator", &coins(2, "token"));
        let _res = instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();

        // decrement the counter
        let _res = handle_callback(deps.as_mut(), 0);
        let res = query(deps.as_ref(), mock_env(), QueryMsg::GetCount {}).unwrap();
        let value: GetCountResponse = from_json(&res).unwrap();
        assert_eq!(99, value.count);

        // increment the counter
        let _res = handle_callback(deps.as_mut(), 1);
        let res = query(deps.as_ref(), mock_env(), QueryMsg::GetCount {}).unwrap();
        let value: GetCountResponse = from_json(&res).unwrap();
        assert_eq!(100, value.count);

        // return error
        let res = handle_callback(deps.as_mut(), 2);
        match res {
            Err(ContractError::SomeError {}) => {}
            _ => panic!("Must return some error"),
        }

        // do nothing
        let _res = handle_callback(deps.as_mut(), 3);
        let res = query(deps.as_ref(), mock_env(), QueryMsg::GetCount {}).unwrap();
        let value: GetCountResponse = from_json(&res).unwrap();
        assert_eq!(100, value.count);
    }

    
    #[test]
    fn error_encountered() {
        let mut deps = mock_dependencies();

        let msg = InstantiateMsg { count: 100 };
        let info = mock_info("creator", &coins(2, "token"));
        let _res = instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();

        let res = query(deps.as_ref(), mock_env(), QueryMsg::GetCount {}).unwrap();
        let value: GetCountResponse = from_json(&res).unwrap();
        assert_eq!(false, value.error_encountered);

        let module_name = "callback".to_string();
        let error_code = 1;
        let input_payload = "input_payload".to_string();
        let error_message = "error_message".to_string();
        let _res = handle_error(deps.as_mut(), module_name, error_code, input_payload, error_message);
        let res = query(deps.as_ref(), mock_env(), QueryMsg::GetCount {}).unwrap();
        let value: GetCountResponse = from_json(&res).unwrap();
        assert_eq!(true, value.error_encountered);

        // return error
        let res = handle_callback(deps.as_mut(), 2);
        match res {
            Err(ContractError::SomeError {}) => {}
            _ => panic!("Must return some error"),
        }

        // do nothing
        let _res = handle_callback(deps.as_mut(), 3);
        let res = query(deps.as_ref(), mock_env(), QueryMsg::GetCount {}).unwrap();
        let value: GetCountResponse = from_json(&res).unwrap();
        assert_eq!(100, value.count);
    }
}
