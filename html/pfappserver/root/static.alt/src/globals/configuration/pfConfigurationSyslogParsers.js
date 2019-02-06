import i18n from '@/utils/locale'
import pfFieldApiMethodParameters from '@/components/pfFieldApiMethodParameters'
import pfFieldRuleSyslogParserRegex from '@/components/pfFieldRuleSyslogParserRegex'
import pfFormFields from '@/components/pfFormFields'
import pfFormInput from '@/components/pfFormInput'
import pfFormRangeToggle from '@/components/pfFormRangeToggle'
import { pfFieldType as fieldType } from '@/globals/pfField'
import {
  pfConfigurationListColumns,
  pfConfigurationListFields
} from '@/globals/configuration/pfConfiguration'
import {
  and,
  not,
  conditional,
  syslogParserExists,
  limitSiblingFields
} from '@/globals/pfValidators'

const {
  maxLength,
  required
} = require('vuelidate/lib/validators')

export const pfConfigurationSyslogParsersListColumns = [
  pfConfigurationListColumns.status,
  { ...pfConfigurationListColumns.id, ...{ label: i18n.t('Detector') } }, // re-label
  { ...pfConfigurationListColumns.type, ...{ label: i18n.t('Type') } }, // re-label
  pfConfigurationListColumns.buttons
]

export const pfConfigurationSyslogParsersListFields = [
  { ...pfConfigurationListFields.id, ...{ text: i18n.t('Detector') } }, // re-text
  pfConfigurationListFields.type
]

export const pfConfigurationSyslogParsersListConfig = (context = {}) => {
  return {
    columns: pfConfigurationSyslogParsersListColumns,
    fields: pfConfigurationSyslogParsersListFields,
    rowClickRoute (item, index) {
      return { name: 'syslogParser', params: { id: item.id } }
    },
    searchPlaceholder: i18n.t('Search by detector or type'),
    searchableOptions: {
      searchApiEndpoint: 'config/syslog_parsers',
      defaultSortKeys: ['id'],
      defaultSearchCondition: {
        op: 'and',
        values: [{
          op: 'or',
          values: [
            { field: 'id', op: 'contains', value: null },
            { field: 'type', op: 'contains', value: null }
          ]
        }]
      },
      defaultRoute: { name: 'syslogParsers' }
    },
    searchableQuickCondition: (quickCondition) => {
      return {
        op: 'and',
        values: [{
          op: 'or',
          values: [
            { field: 'id', op: 'contains', value: quickCondition },
            { field: 'type', op: 'contains', value: quickCondition }
          ]
        }]
      }
    }
  }
}

export const pfConfigurationSyslogParserRegexRuleActions = {
  add_person: {
    value: 'add_person',
    text: i18n.t('Create new user account'),
    types: [fieldType.SUBSTRING],
    defaultApiParameters: 'pid, $pid',
    validators: {
      api_parameters: {
        [i18n.t('API Parameters required.')]: required
      }
    }
  },
  close_violation: {
    value: 'close_violation',
    text: i18n.t('Close violation'),
    types: [fieldType.SUBSTRING],
    defaultApiParameters: 'mac, $mac, vid, VID',
    validators: {
      api_parameters: {
        [i18n.t('API Parameters required.')]: required
      }
    }
  },
  deregister_node_ip: {
    value: 'deregister_node_ip',
    text: i18n.t('Deregister node by IP'),
    types: [fieldType.SUBSTRING],
    defaultApiParameters: 'ip, $ip',
    validators: {
      api_parameters: {
        [i18n.t('API Parameters required.')]: required
      }
    }
  },
  dynamic_register_node: {
    value: 'dynamic_register_node',
    text: i18n.t('Register node by MAC'),
    types: [fieldType.SUBSTRING],
    defaultApiParameters: 'mac, $mac, username, $username',
    validators: {
      api_parameters: {
        [i18n.t('API Parameters required.')]: required
      }
    }
  },
  modify_node: {
    value: 'modify_node',
    text: i18n.t('Modify node by MAC'),
    types: [fieldType.SUBSTRING],
    defaultApiParameters: 'mac, $mac',
    validators: {
      api_parameters: {
        [i18n.t('API Parameters required.')]: required
      }
    }
  },
  modify_person: {
    value: 'modify_person',
    text: i18n.t('Modify existing user'),
    types: [fieldType.SUBSTRING],
    defaultApiParameters: 'pid, $pid',
    validators: {
      api_parameters: {
        [i18n.t('API Parameters required.')]: required
      }
    }
  },
  reevaluate_access: {
    value: 'reevaluate_access',
    text: i18n.t('Reevaluate access by MAC'),
    types: [fieldType.SUBSTRING],
    defaultApiParameters: 'mac, $mac, reason, $reason',
    validators: {
      api_parameters: {
        [i18n.t('API Parameters required.')]: required
      }
    }
  },
  register_node: {
    value: 'register_node',
    text: i18n.t('Register a new node by PID'),
    types: [fieldType.SUBSTRING],
    defaultApiParameters: 'mac, $mac, pid, $pid',
    validators: {
      api_parameters: {
        [i18n.t('API Parameters required.')]: required
      }
    }
  },
  register_node_ip: {
    value: 'register_node_ip',
    text: i18n.t('Register node by IP'),
    types: [fieldType.SUBSTRING],
    defaultApiParameters: 'ip, $ip, pid, $pid',
    validators: {
      api_parameters: {
        [i18n.t('API Parameters required.')]: required
      }
    }
  },
  release_all_violations: {
    value: 'release_all_violations',
    text: i18n.t('Release all violations for node by MAC'),
    types: [fieldType.SUBSTRING],
    defaultApiParameters: '$mac',
    validators: {
      api_parameters: {
        [i18n.t('API Parameters required.')]: required
      }
    }
  },
  role_detail: {
    value: 'role_detail',
    text: i18n.t('role_detail'),
    types: [fieldType.SUBSTRING],
    defaultApiParameters: 'role, $role',
    validators: {
      api_parameters: {
        [i18n.t('API Parameters required.')]: required
      }
    }
  },
  trigger_scan: {
    value: 'trigger_scan',
    text: i18n.t('Launch a scan for the device'),
    types: [fieldType.SUBSTRING],
    defaultApiParameters: '$ip, mac, $mac, net_type, TYPE',
    validators: {
      api_parameters: {
        [i18n.t('API Parameters required.')]: required
      }
    }
  },
  trigger_violation: {
    value: 'trigger_violation',
    text: i18n.t('Trigger a violation'),
    types: [fieldType.SUBSTRING],
    defaultApiParameters: 'mac, $mac, tid, TYPEID, type, TYPE',
    validators: {
      api_parameters: {
        [i18n.t('API Parameters required.')]: required
      }
    }
  },
  unreg_node_for_pid: {
    value: 'unreg_node_for_pid',
    text: i18n.t('Deregister node by PID'),
    types: [fieldType.SUBSTRING],
    defaultApiParameters: 'pid, $pid',
    validators: {
      api_parameters: {
        [i18n.t('API Parameters required.')]: required
      }
    }
  },
  update_ip4log: {
    value: 'update_ip4log',
    text: i18n.t('Update ip4log by IP and MAC'),
    types: [fieldType.SUBSTRING],
    defaultApiParameters: 'mac, $mac, ip, $ip',
    validators: {
      api_parameters: {
        [i18n.t('API Parameters required.')]: required
      }
    }
  },
  update_ip6log: {
    value: 'update_ip6log',
    text: i18n.t('Update ip6log by IP and MAC'),
    types: [fieldType.SUBSTRING],
    defaultApiParameters: 'mac, $mac, ip, $ip',
    validators: {
      api_parameters: {
        [i18n.t('API Parameters required.')]: required
      }
    }
  },
  update_role_configuration: {
    value: 'update_role_configuration',
    text: i18n.t('Update role configuration'),
    types: [fieldType.SUBSTRING],
    defaultApiParameters: 'role, $role',
    validators: {
      api_parameters: {
        [i18n.t('API Parameters required.')]: required
      }
    }
  }
}

export const pfConfigurationSyslogParserViewFields = (context) => {
  const {
    isNew = false,
    isClone = false,
    syslogParserType = null
  } = context
  return [
    {
      tab: null, // ignore tabs
      fields: [
        {
          label: i18n.t('Detector'),
          fields: [
            {
              key: 'id',
              component: pfFormInput,
              attrs: {
                disabled: (!isNew && !isClone)
              },
              validators: {
                [i18n.t('Value required.')]: required,
                [i18n.t('Maximum 255 characters.')]: maxLength(255),
                [i18n.t('Syslog Parser exists.')]: not(and(required, conditional(isNew || isClone), syslogParserExists))
              }
            }
          ]
        },
        {
          label: i18n.t('Enabled'),
          fields: [
            {
              key: 'status',
              component: pfFormRangeToggle,
              attrs: {
                values: { checked: 'enabled', unchecked: 'disabled' }
              }
            }
          ]
        },
        {
          label: i18n.t('Alert pipe'),
          fields: [
            {
              key: 'path',
              component: pfFormInput,
              validators: {
                [i18n.t('Value required.')]: required,
                [i18n.t('Maximum 255 characters.')]: maxLength(255)
              }
            }
          ]
        },
        {
          if: ['regex'].includes(syslogParserType), // 'regex' only
          label: 'Rules',
          fields: [
            {
              key: 'rules',
              component: pfFormFields,
              attrs: {
                buttonLabel: 'Add Rule - New ( )',
                sortable: true,
                field: {
                  component: pfFieldRuleSyslogParserRegex,
                  attrs: {
                    actions: {
                      component: pfFieldApiMethodParameters,
                      attrs: {
                        typeLabel: i18n.t('Select action type'),
                        valueLabel: i18n.t('Select action value'),
                        fields: [
                          pfConfigurationSyslogParserRegexRuleActions.deregister_node_ip,
                          pfConfigurationSyslogParserRegexRuleActions.role_detail,
                          pfConfigurationSyslogParserRegexRuleActions.modify_person,
                          pfConfigurationSyslogParserRegexRuleActions.register_node_ip,
                          pfConfigurationSyslogParserRegexRuleActions.update_ip6log,
                          pfConfigurationSyslogParserRegexRuleActions.add_person,
                          pfConfigurationSyslogParserRegexRuleActions.unreg_node_for_pid,
                          pfConfigurationSyslogParserRegexRuleActions.trigger_scan,
                          pfConfigurationSyslogParserRegexRuleActions.release_all_violations,
                          pfConfigurationSyslogParserRegexRuleActions.close_violation,
                          pfConfigurationSyslogParserRegexRuleActions.reevaluate_access,
                          pfConfigurationSyslogParserRegexRuleActions.update_ip4log,
                          pfConfigurationSyslogParserRegexRuleActions.update_role_configuration,
                          pfConfigurationSyslogParserRegexRuleActions.trigger_violation,
                          pfConfigurationSyslogParserRegexRuleActions.modify_node,
                          pfConfigurationSyslogParserRegexRuleActions.register_node,
                          pfConfigurationSyslogParserRegexRuleActions.dynamic_register_node
                        ]
                      },
                      invalidFeedback: [
                        { [i18n.t('Action(s) contain one or more errors.')]: true }
                      ]
                    }
                  },
                  validators: {
                    id: {
                      [i18n.t('Name required.')]: required,
                      [i18n.t('Maximum 255 characters.')]: maxLength(255),
                      [i18n.t('Duplicate name.')]: limitSiblingFields('id', 0)
                    },
                    regex: {
                      [i18n.t('Regex required.')]: required,
                      [i18n.t('Maximum 255 characters.')]: maxLength(255)
                    }
                  }
                },
                invalidFeedback: [
                  { [i18n.t('Rule(s) contain one or more errors.')]: true }
                ]
              }
            }
          ]
        }
      ]
    }
  ]
}

export const pfConfigurationSyslogParserViewDefaults = (context = {}) => {
  const {
    syslogParserType = null
  } = context
  switch (syslogParserType) {
    default:
      return {}
  }
}
