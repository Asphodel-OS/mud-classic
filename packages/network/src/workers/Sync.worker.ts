import { runWorker } from "@mud-classic/utils";
import { SyncWorker } from "./SyncWorker";

runWorker(new SyncWorker());
