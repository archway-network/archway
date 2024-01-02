use cosmwasm_std::{DepsMut, Empty, Env, MessageInfo, Response, entry_point};
use crate::errors::ContractError;
use crate::msgs::{CWGrant, InstantiateMsg, SudoMsg};
use crate::state::GRANTS;

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    for addr in &msg.grants {
        let addr = deps.api.addr_validate(addr)?;
        GRANTS.save(deps.storage, &addr, &Empty{})?
    }
    Ok(Response::default())
}
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn sudo(
    deps: DepsMut,
    _env: Env,
    msg: SudoMsg,
) -> Result<Response, ContractError> {
    return match msg {
        SudoMsg::CWGrant(grant) => {
            sudo_grant(deps, grant)
        }
    }
}

fn sudo_grant(deps: DepsMut, msg: CWGrant) -> Result<Response, ContractError> {
    // in order to pay the fees all message senders need to be
    // in the grants list.
    for m in &msg.msgs {
        let sender = deps.api.addr_validate(&m.sender)?;
        if !GRANTS.has(deps.storage, &sender) {
            return Err(ContractError::Unauthorized {})
        }
    }

    Ok(Response::default())
}
