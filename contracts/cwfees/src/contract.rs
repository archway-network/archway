use cosmwasm_std::{DepsMut, Empty, Env, MessageInfo, Response, entry_point};
use crate::errors::ContractError;
use crate::msgs::{CwGrant, InstantiateMsg, SudoMsg};
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
        SudoMsg::CwGrant(grant) => {
            sudo_grant(deps, grant)
        }
    }
}

fn sudo_grant(deps: DepsMut, msg: CwGrant) -> Result<Response, ContractError> {
    // add a malicious case in which if the fee request contains a hack denom
    // then we basically compute forever.
    for fee in &msg.fee_requested {
        if fee.denom == "malicious" {
            let mut x = 0;
            loop {
                x+=1
            }
        }
    }
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

#[cfg(test)]
mod test {
    use cosmwasm_std::to_json_binary;
    use crate::msgs::{CwGrant, SudoMsg};

    #[test]
    fn encoding() {
        println!("{}", to_json_binary(&SudoMsg::CwGrant(CwGrant{
            fee_requested: vec![],
            msgs: vec![],
        })).unwrap());
    }
}
