import { Components } from "@mud-classic/recs";
import { fromWorker } from "@mud-classic/utils";
import { map, Observable, Subject, timer } from "rxjs";
import { NetworkEvent } from "./types";
import { Input, Ack, ack } from "./workers/SyncWorker";

/**
 * Create a new SyncWorker ({@link Sync.worker.ts}) to performn contract/client state sync.
 * The main thread and worker communicate via RxJS streams.
 *
 * @returns Object {
 * ecsEvent$: Stream of network component updates synced by the SyncWorker,
 * config$: RxJS subject to pass in config for the SyncWorker,
 * dispose: function to dispose of the sync worker
 * }
 */

export function createSyncWorker<C extends Components = Components>(ack$?: Observable<Ack>) {
  const workerCode = `
    import { runWorker } from "@mud-classic/utils";
    import { SyncWorker } from "./SyncWorker";
    runWorker(new SyncWorker());
 `;

  const blob = new Blob([workerCode], { type: 'text/javascript' });
  const url = URL.createObjectURL(blob);
  const worker = new Worker(url, { type: 'module' });

  // Send ack every 16ms if no external ack$ is provided
  const input$ = new Subject<Input>();
  ack$ = ack$ || timer(0, 16).pipe(map(() => ack));
  const ackSub = ack$.subscribe(input$);

  // Pass in a "config stream", receive a stream of ECS events
  const ecsEvents$ = new Subject<NetworkEvent<C>[]>();
  const subscription = fromWorker<Input, NetworkEvent<C>[]>(worker, input$).subscribe(ecsEvents$);
  const dispose = () => {
    worker.terminate();
    subscription?.unsubscribe();
    ackSub?.unsubscribe();
  };

  return {
    ecsEvents$,
    input$,
    dispose,
  };
}