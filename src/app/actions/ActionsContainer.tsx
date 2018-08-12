import * as React from 'react'
import { Subscription } from 'rxjs'
import { ConfigurationSubject, Settings } from '../../settings'
import { ActionItem, ActionItemProps } from './ActionItem'
import { ActionsProps, ActionsState } from './actions'
import { getContributedActionItems } from './contributions'

interface ActionsContainerProps<S extends ConfigurationSubject, C extends Settings> extends ActionsProps<S, C> {
    /**
     * Called with the array of contributed items to produce the rendered component. If not set, uses a default
     * render function that renders a <ActionItem> for each item.
     */
    render?: (items: ActionItemProps[]) => React.ReactElement<any>

    /**
     * If set, it is rendered when there are no contributed items for this menu. Use null to render nothing when
     * empty.
     */
    empty?: React.ReactElement<any> | null
}

interface ActionsContainerState extends ActionsState {}

/** Displays the actions in a container, with a wrapper and/or empty element. */
export class ActionsContainer<S extends ConfigurationSubject, C extends Settings> extends React.PureComponent<
    ActionsContainerProps<S, C>,
    ActionsContainerState
> {
    public state: ActionsState = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.props.cxpController.registries.contribution.contributions.subscribe(contributions =>
                this.setState({ contributions })
            )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state.contributions) {
            return null // loading
        }

        const items = getContributedActionItems(this.state.contributions, this.props.menu)
        if (this.props.empty !== undefined && items.length === 0) {
            return this.props.empty
        }

        const render = this.props.render || this.defaultRenderItems
        return render(items)
    }

    private defaultRenderItems = (items: ActionItemProps[]): JSX.Element | null => (
        <>
            {items.map((item, i) => (
                <ActionItem
                    key={i}
                    {...item}
                    cxpController={this.props.cxpController}
                    extensions={this.props.extensions}
                />
            ))}
        </>
    )
}
